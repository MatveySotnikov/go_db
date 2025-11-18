# ПР5 - Подключение к PostgreSQL через database/sql. Выполнение простых запросов (INSERT, SELECT)
## Выполнил Сотников М.Е. ЭФМО-01-25

Создание базы данных через psql:  

<img width="612" height="373" alt="СозданиеБД" src="https://github.com/user-attachments/assets/2a7518b5-e101-4e60-bbdb-3c1f4f1cf318" />

Запуск программы, в которой добавляются 3 задачи:

<img width="695" height="268" alt="ЗапускGo" src="https://github.com/user-attachments/assets/2a86b275-9f09-4c5b-91d3-e024674ce0ad" />   

Проверка наличия задач в БД через psql:   

<img width="614" height="303" alt="Добавленные задачи" src="https://github.com/user-attachments/assets/f4f954fb-a41c-44ac-93e3-19f2f533b024" />

## Функция ListDone()

ListDone возвращает срез задач, отфильтрованный по их статусу выполнения. Он принимает булево значение done (true для выполненных, false для невыполненных) и контекст. Использует параметризованный запрос (WHERE done = $1) для безопасного взаимодействия с базой данных. Если задачи не найдены, возвращается пустой срез.   

```golang
func (r *Repo) ListDone(ctx context.Context, done bool) ([]Task, error) {
	const q = `SELECT id, title, done, created_at FROM tasks WHERE done = $1 ORDER BY id;`

	rows, err := r.DB.QueryContext(ctx, q, done) // используем $1 = 'done' для предотвращения инъекций
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Task
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.Title, &t.Done, &t.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}
```

<img width="689" height="427" alt="ListDoneCheck" src="https://github.com/user-attachments/assets/7578b19a-6e95-488d-a819-7e3092d9ce5b" />   

## Функция FindById()   

Метод FindByID осуществляет поиск и возвращает одну задачу (*Task) по ее уникальному id. Метод корректно обрабатывает отсутствие задачи в базе данных, возвращая nil вместо ошибки sql.ErrNoRows. Это позволяет вызывающему коду легко определить, была ли задача найдена.

```golang
func (r *Repo) FindByID(ctx context.Context, id int) (*Task, error) {
	const q = `SELECT id, title, done, created_at FROM tasks WHERE id = $1;`
	var t Task

	err := r.DB.QueryRowContext(ctx, q, id).Scan(&t.ID, &t.Title, &t.Done, &t.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}
```

## CreateMany
Метод CreateMany предназначен для выполнения массовой вставки новых задач в базу данных. В функции используются транзакции, поэтому либо все задачи из переданного списка будут успешно вставлены в таблицу, либо, в случае любой ошибки, все изменения будут откачены (Rollback), и ни одна задача не попадет в базу данных. Для повышения производительности внутри транзакции используется подготовленный запрос.    

```golang
func (r *Repo) CreateMany(ctx context.Context, titles []string) error {
	tx, err := r.DB.BeginTx(ctx, nil) //начало транзакции
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback() //откат при ошибке
		}
	}()

	const q = `INSERT INTO tasks (title) VALUES ($1);`

	stmt, err := tx.PrepareContext(ctx, q) //подготовка запроса для многократного использования
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, title := range titles {
		_, err = stmt.ExecContext(ctx, title) //выполнение подготовленного запроса
		if err != nil {
			return err
		}
	}

	err = tx.Commit() //фиксация транзакции
	return err
}
```

<img width="652" height="314" alt="УспешноеCreateMany" src="https://github.com/user-attachments/assets/5b690ad0-525a-4f34-94f7-866dcf1517c0" />





