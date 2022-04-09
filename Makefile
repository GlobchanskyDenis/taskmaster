UNIT_BIN			=	unit_bin
UNIT_DIR			=	cmd/unit
UNIT_FILES		=	main.go		flags.go	config.go	debug.go
UNIT_FILENAMES	=	$(addprefix $(UNIT_DIR)/,$(UNIT_FILES))

SAMPLE_SIMPLE_BIN		=	sample_simple_bin
SAMPLE_SIMPLE_DIR		=	cmd/sample_simple
SAMPLE_SIMPLE_FILES		=	main.go
SAMPLE_SIMPLE_FILENAMES	=	$(addprefix $(SAMPLE_SIMPLE_DIR)/,$(SAMPLE_SIMPLE_FILES))

SAMPLE_AUTORESTART_BIN			=	sample_autorestart_bin
SAMPLE_AUTORESTART_DIR			=	cmd/sample_autorestart
SAMPLE_AUTORESTART_FILES		=	main.go
SAMPLE_AUTORESTART_FILENAMES	=	$(addprefix $(SAMPLE_AUTORESTART_DIR)/,$(SAMPLE_AUTORESTART_FILES))

SAMPLE_ARGS_ENV_BIN				=	sample_args_env_bin
SAMPLE_ARGS_ENV_DIR				=	cmd/sample_args_env
SAMPLE_ARGS_ENV_FILES			=	main.go		flags.go		env.go	umask_dir.go
SAMPLE_ARGS_ENV_FILENAMES		=	$(addprefix $(SAMPLE_ARGS_ENV_DIR)/,$(SAMPLE_ARGS_ENV_FILES))

SAMPLE_SOCKET_CLIENT_BIN		=	sample_client
SAMPLE_SOCKET_CLIENT_DIR		=	cmd/sample_socket_client
SAMPLE_SOCKET_CLIENT_FILES		=	main.go
SAMPLE_SOCKET_CLIENT_FILENAMES	=	$(addprefix $(SAMPLE_SOCKET_CLIENT_DIR)/,$(SAMPLE_SOCKET_CLIENT_FILES))

SAMPLE_SOCKET_SERVER_BIN		=	sample_server
SAMPLE_SOCKET_SERVER_DIR		=	cmd/sample_socket_server
SAMPLE_SOCKET_SERVER_FILES		=	main.go
SAMPLE_SOCKET_SERVER_FILENAMES	=	$(addprefix $(SAMPLE_SOCKET_SERVER_DIR)/,$(SAMPLE_SOCKET_SERVER_FILES))

all : $(UNIT_BIN) $(SAMPLE_SIMPLE_BIN) $(SAMPLE_AUTORESTART_BIN) $(SAMPLE_ARGS_ENV_BIN) $(SAMPLE_SOCKET_CLIENT_BIN) $(SAMPLE_SOCKET_SERVER_BIN)

$(UNIT_BIN) : $(UNIT_FILENAMES)
	@echo "компилирую бинарник супервизора"
	@go build -o $(UNIT_BIN) $(UNIT_FILENAMES)

$(SAMPLE_SIMPLE_BIN) : $(SAMPLE_SIMPLE_FILENAMES)
	@echo "компилирую бинарник примера 1 (hello world)"
	@go build -o $(SAMPLE_SIMPLE_BIN) $(SAMPLE_SIMPLE_FILENAMES)

$(SAMPLE_AUTORESTART_BIN) : $(SAMPLE_AUTORESTART_FILENAMES)
	@echo "компилирую бинарник примера 2 (авторестарт)"
	@go build -o $(SAMPLE_AUTORESTART_BIN) $(SAMPLE_AUTORESTART_FILENAMES)

$(SAMPLE_ARGS_ENV_BIN) : $(SAMPLE_ARGS_ENV_FILENAMES)
	@echo "компилирую бинарник примера 3 (флаги, переменные окружения)"
	@go build -o $(SAMPLE_ARGS_ENV_BIN) $(SAMPLE_ARGS_ENV_FILENAMES)

$(SAMPLE_SOCKET_CLIENT_BIN) : $(SAMPLE_SOCKET_CLIENT_FILENAMES)
	@echo "компилирую бинарник примера 4.1 (клиент работы с сокетом)"
	@go build -o $(SAMPLE_SOCKET_CLIENT_BIN) $(SAMPLE_SOCKET_CLIENT_FILENAMES)

$(SAMPLE_SOCKET_SERVER_BIN) : $(SAMPLE_SOCKET_SERVER_FILENAMES)
	@echo "компилирую бинарник примера 4.2 (сервер работы с сокетом)"
	@go build -o $(SAMPLE_SOCKET_SERVER_BIN) $(SAMPLE_SOCKET_SERVER_FILENAMES)

fclean:
	rm -rf $(UNIT_BIN)
	rm -rf $(SAMPLE_SIMPLE_BIN)
	rm -rf $(SAMPLE_AUTORESTART_BIN)
	rm -rf $(SAMPLE_ARGS_ENV_BIN)
	rm -rf $(SAMPLE_SOCKET_CLIENT_BIN)
	rm -rf $(SAMPLE_SOCKET_SERVER_BIN)

re: fclean all