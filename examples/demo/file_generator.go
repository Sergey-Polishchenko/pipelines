package main

import (
	"context"
	"os"
	"path/filepath"

	"github.com/Sergey-Polishchenko/pipelines/nodes"
)

// FileGenerator возвращает функцию-генератор,
// которая при каждом вызове отдаёт следующий путь файла из rootDir.
// Когда файлы заканчиваются — возвращает пустую строку.
func FileGenerator(rootDir string) nodes.Generator[string] {
	return func(ctx context.Context) (<-chan string, error) {
		out := make(chan string)
		go func() {
			defer close(out)
			// Вместо накопления всех файлов в слайс, сразу стримим результаты:
			filepath.WalkDir(rootDir, func(path string, d os.DirEntry, err error) error {
				if err != nil || d.IsDir() {
					return nil
				}
				select {
				case out <- path:
					return nil
				case <-ctx.Done():
					return ctx.Err()
				}
			})
		}()
		return out, nil
	}
}
