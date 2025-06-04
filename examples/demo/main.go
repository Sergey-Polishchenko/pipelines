package main

import (
	"context"
	"crypto/md5"
	"fmt"
	"os"
	"time"

	"github.com/Sergey-Polishchenko/pipelines"
	"github.com/Sergey-Polishchenko/pipelines/nodes"
)

type FileHash struct {
	Path string
	Hash [16]byte
}

func main() {
	// контекст с таймаутом (например, 5 минут), но вы можете убрать таймаут, если он не нужен
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// 1) Генератор, который выдаёт пути файлов
	fileGen := nodes.NewGenerator(FileGenerator("./"))

	// 2) Пул воркеров: он сам распараллеливает MD5-вычисления.
	//    По умолчанию DefaultConfig() содержит Workers=10.
	//    Если хотите другой параллелизм, передайте Config{Workers: N}.
	md5Worker := nodes.NewWorkerPool(func(path string) (FileHash, error) {
		data, err := os.ReadFile(path)
		if err != nil {
			return FileHash{}, err
		}
		return FileHash{
			Path: path,
			Hash: md5.Sum(data),
		}, nil
	}, nodes.DefaultConfig())

	// 3) Агрегатор итогов: можно либо просто печатать, либо накапливать в срез для дальнейшей обработки.
	//    Здесь для демонстрации печатаем в stdout.
	resultAgg := nodes.NewResultAggregator(func(fh FileHash) error {
		fmt.Printf("%x  %s\n", fh.Hash, fh.Path)
		return nil
	})

	// Составляем пайплайн: fileGen -> md5Worker -> resultAgg
	p := pipelines.New()
	p.Add(fileGen, md5Worker, resultAgg)

	// Соединяем их:
	//   fileGen.Output() -> md5Worker.SetInput()
	//   md5Worker.Output() -> resultAgg.SetInput()
	// (в однопоточном конвейере Connect делает всё, что нужно)
	if err := pipelines.Connect(fileGen, md5Worker); err != nil {
		fmt.Fprintf(os.Stderr, "Connect fileGen -> md5Worker: %v\n", err)
		return
	}
	if err := pipelines.Connect(md5Worker, resultAgg); err != nil {
		fmt.Fprintf(os.Stderr, "Connect md5Worker -> resultAgg: %v\n", err)
		return
	}

	// Запускаем весь пайплайн
	if err := p.Run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Pipeline error: %v\n", err)
	}
}
