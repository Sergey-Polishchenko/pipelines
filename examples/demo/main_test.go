package main_test

import (
	"context"
	"crypto/md5"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Sergey-Polishchenko/pipelines"
	"github.com/Sergey-Polishchenko/pipelines/nodes"

	mainpkg "github.com/Sergey-Polishchenko/pipelines/examples/demo"
)

func TestPipeline(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	// 1. Создаём временную директорию и пару файлов внутри
	dir := t.TempDir()
	files := map[string]string{
		"file1.txt": "hello",
		"file2.txt": "world",
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
			t.Fatalf("cannot write file %s: %v", name, err)
		}
	}

	// 2. Генератор, который отдаёт пути найденных файлов
	fileGen := nodes.NewGenerator(mainpkg.FileGenerator(dir))

	// 3. Пул воркеров, который читает файл и считает MD5
	workers := nodes.NewWorkerPool(func(path string) (mainpkg.FileHash, error) {
		data, err := os.ReadFile(path)
		return mainpkg.FileHash{
			Path: path,
			Hash: md5.Sum(data),
		}, err
	}, nodes.Config{Workers: 2})

	// 4. Нода-агрегатор, которая собирает результаты в срез
	var results []mainpkg.FileHash
	resultsNode := nodes.NewResultAggregator(func(fh mainpkg.FileHash) error {
		results = append(results, fh)
		return nil
	})

	// 5. Соединяем: fileGen → workers → resultsNode
	if err := pipelines.Connect(fileGen, workers); err != nil {
		t.Fatalf("Connect(fileGen, workers) failed: %v", err)
	}
	if err := pipelines.Connect(workers, resultsNode); err != nil {
		t.Fatalf("Connect(workers, resultsNode) failed: %v", err)
	}

	// 6. Создаём и запускаем с��м Pipeline
	pipeline := pipelines.New()
	pipeline.Add(fileGen, workers, resultsNode)
	if err := pipeline.Run(ctx); err != nil {
		t.Fatal("Pipeline error:", err)
	}

	// 7. Проверяем, что мы получили ровно два результата и они совпадают с ожидаемыми
	expected := map[string][16]byte{}
	for name, content := range files {
		expected[filepath.Join(dir, name)] = md5.Sum([]byte(content))
	}

	for _, fh := range results {
		if hash, ok := expected[fh.Path]; !ok || hash != fh.Hash {
			t.Errorf("Unexpected hash or file: got %s %x", fh.Path, fh.Hash)
		}
		delete(expected, fh.Path)
	}
	if len(expected) != 0 {
		t.Errorf("Some files were not processed: %v", expected)
	}
}
