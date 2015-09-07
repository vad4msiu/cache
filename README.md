### Задание
Необходимо реализовать package для распределенного кеша приложения.

Интерфейс: middleware с передачей анонимной функции для получения данных:

func Get(key string, result interface{}, executionLimit, expire time.Duration, getter func()) (error)

Данные представляют собой нормально сериализуемые json структуры.
Сложность в том, чтобы в нормальных условиях гарантировать уникальность вызова функции в
распределённой системе в пределах expire (т.е. все прочие вызовы должны запросить кеш).

Поскольку это тестовое задание, ограничений в использовании сторонних пакетов и баз данных нет (см. ключевую метрику оценки).
Неэффективная обработка редких проблем так же несущественна. Важным плюсом будет написанный бенчмарк, демонстрирующий производительность решения.

Масштабы: суммарно данных для кеша очень много (текущий срез сравним с суммарной памятью доступных серверов), разные ключи запрашиваются очень неравномерно (распределение Ципфа).

Ключевая метрика оценки решения: скорость ответа в нормальной ситуации и пропускная способность системы, всё остальное вторично. Особым плюсом будет архитектурное решение, надёжно работающее в условиях падения серверов.

### Решение
За основу было взято key-value хранилище Aerospike.

Уникальность вызова функции достигается за счет атомарной операции PutBins с политикой CREATE_ONLY.
То есть если 10 конкурирующих поток попытаются создать запись с ключом "SOME_KEY", то 9 из них получат ошибку.
Тот поток который успешно создал ключ "SOME_KEY" получает блокировку на эту запись и вызывает функцию getter для получения и записи данных.
Остальные 9 потоков ожидают появления данных в этом ключе (не больше executionLimit времени).

Хранилище Aerospike очень фиче развесистое.
Достаточно просто развертывается кластер с репликациями, что позволит надежно работать при отказе какого либо из узлов.

Для теста был создан инстанс Aerospike на google cloud n1-standard-4 (4 vCPUs, 15 GB memory).
Результаты бенчмарка для 10 000 000 вызовов func Get(...)
![img 1](https://github.com/vad4msiu/cache/blob/master/imgs/9sths3pj.png)
Пик операций на записей был порядка 45 000 в сек.
![img 2](https://github.com/vad4msiu/cache/blob/master/imgs/7h2337hi.png)
На чтение 53 000 в сек.
![img 3](https://github.com/vad4msiu/cache/blob/master/imgs/-16iusyg.png)
