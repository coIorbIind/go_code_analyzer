# Go Code Analyzer

## Логика работы

---

## Структура проекта
````
.
├── cmd
|   └── analyzer — запуск проекта
├── internal
|   └── graph_builder — логика построения графа вызовов
|       ├── graph.go — содержит собственно сам граф и его методы
|       └── parser.go — содержит парсинг и обработку всего проекта
├── .gitignore
├── go.mod
├── go.sum
└── README.md
````
---

## Аргументы запуска
* `-exclude=` — список пакетов, которые не нужно обрабатывать
* `-include=` — список пакетов, которые нужно обрабатывать

### Примеры запуска
Выполните сборку проекта командой `go build -o callgraph cmd/analyzer/main.go`

* Вывод графа вызовов в терминал: `./callgraph ./path/to/project`.
* Сохранение графа вызовов в `.dot` файл: `./callgraph ./path/to/project > graph.dot`.
* Сохранение графа вызовов в `png` формате: `./callgraph ./path/to/project | dot -Tpng -o callgraph.png` 
(необходима утилита Graphviz)
* 

