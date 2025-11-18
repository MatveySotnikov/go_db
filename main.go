package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	// .env не обязателен; если файла нет — ошибка игнорируется
	_ = godotenv.Load()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		// fallback — прямой DSN в коде (только для учебного стенда!)
		dsn = "postgres://postgres:YOURPASSWORD@localhost:5432/todo?sslmode=disable"
	}

	db, err := openDB(dsn)
	if err != nil {
		log.Fatalf("openDB error: %v", err)
	}
	defer db.Close()

	repo := NewRepo(db)

	// 1) Вставим пару задач
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	titles := []string{"Без транзакции1", "Без транзакции2", "Без транзакции3"}
	for _, title := range titles {
		id, err := repo.CreateTask(ctx, title)
		if err != nil {
			log.Fatalf("CreateTask error: %v", err)
		}
		log.Printf("Inserted task id=%d (%s)", id, title)
	}

	manyTitles := []string{"Задача 1 (Транзакция)", "Задача 2 (Транзакция)", "Задача 3 (Транзакция)"}
	ctxMany, cancelMany := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelMany()

	log.Printf("Запуск CreatMany для %d задач...", len(manyTitles))
	if err := repo.CreateMany(ctxMany, manyTitles); err != nil {
		log.Fatalf("Ошибка: %v", err)
	}
	log.Printf("Добавлены %d задач через транзакции.", len(manyTitles))

	// Прочитаем список задач
	ctxList, cancelList := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancelList()

	tasks, err := repo.ListTasks(ctxList)
	if err != nil {
		log.Fatalf("ListTasks error: %v", err)
	}

	// Напечатаем
	fmt.Println("=== Tasks ===")
	for _, t := range tasks {
		fmt.Printf("#%d | %-24s | done=%-5v | %s\n",
			t.ID, t.Title, t.Done, t.CreatedAt.Format(time.RFC3339))
	}

	// Проверка ListDont(false)
	ctxUndone, cancelUndone := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancelUndone()

	undoneTasks, err := repo.ListDone(ctxUndone, false)
	if err != nil {
		log.Fatalf("ListDone error (undone): %v", err)
	}

	fmt.Println("\n=== Undone Tasks (ListDone(false)) ===")
	for _, t := range undoneTasks {
		fmt.Printf("#%d | %-24s | done=%-5v | %s\n",
			t.ID, t.Title, t.Done, t.CreatedAt.Format(time.RFC3339))
	}

	ctxDoneList, cancelDoneList := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancelDoneList()

	// Проверка ListDone(true)
	doneTasks, err := repo.ListDone(ctxDoneList, true)
	if err != nil {
		log.Fatalf("ListDone error (done): %v", err)
	}

	fmt.Println("\n=== Done Tasks (ListDone(true)) ===")
	for _, t := range doneTasks {
		fmt.Printf("#%d | %-24s | done=%-5v | %s\n",
			t.ID, t.Title, t.Done, t.CreatedAt.Format(time.RFC3339))
	}

	// Проверка FindById
	targetID := 29
	ctxFind, cancelFind := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancelFind()

	taskDetails, err := repo.FindByID(ctxFind, targetID)
	if err != nil {
		log.Fatalf("FindByID error: %v", err)
	}

	fmt.Printf("\n=== Task Details (ID=%d) ===\n", targetID)
	if taskDetails != nil {
		fmt.Printf("#%d | %-24s | done=%-5v | %s\n",
			taskDetails.ID, taskDetails.Title, taskDetails.Done, taskDetails.CreatedAt.Format(time.RFC3339))
	} else {
		fmt.Printf("Task with ID=%d not found.\n", targetID)
	}

}
