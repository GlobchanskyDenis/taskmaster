# Описание

Pet проект - програмка функционалом напоминает systemctl. Запускает процессы, следит за ними, позволяет останавливать, стартовать, рестартовать их и проверяет их статусы. Исполняется в виде двух программ - первая супервизор, вторая - терминал для работы. Они связываются друг с другом посредством стистемного сокета.

В последствии возможно допишу дополнительные "фронты" для для управления процессами - веб морду и оконное приложение