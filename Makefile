SOCKET_BIN			=	socket_bin
SOCKET_DIR			=	cmd/socket
SOCKET_FILES		=	main.go		flags.go	config.go
SOCKET_FILENAMES	=	$(addprefix $(SOCKET_DIR)/,$(SOCKET_FILES))

SAMPLE_SIMPLE_BIN		=	sample_simple_bin
SAMPLE_SIMPLE_DIR		=	cmd/sample_simple
SAMPLE_SIMPLE_FILES		=	main.go
SAMPLE_SIMPLE_FILENAMES	=	$(addprefix $(SAMPLE_SIMPLE_DIR)/,$(SAMPLE_SIMPLE_FILES))


all : $(SOCKET_BIN) $(SAMPLE_SIMPLE_BIN)

$(SOCKET_BIN) : $(SOCKET_FILENAMES)
	@echo "компилирую бинарник супервизора"
	@go build -o $(SOCKET_BIN) $(SOCKET_FILENAMES)

$(SAMPLE_SIMPLE_BIN) : $(SAMPLE_SIMPLE_FILENAMES)
	@echo "компилирую бинарник примера 1 (hello world)"
	@go build -o $(SAMPLE_SIMPLE_BIN) $(SAMPLE_SIMPLE_FILENAMES)

fclean:
	rm -rf $(SOCKET_BIN)
	rm -rf $(SAMPLE_SIMPLE_BIN)

re: fclean all