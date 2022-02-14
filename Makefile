SOCKET_BIN		=	socket
SOCKET_DIR		=	cmd/socket
SOCKET_FILES		=	main.go		config.go
SOCKET_FILENAMES	=	$(addprefix $(SOCKET_DIR)/,$(SOCKET_FILES))

all : $(SOCKET_BIN)

$(SOCKET_BIN) : $(SOCKET_FILENAMES)
	@echo "компилирую все в бинарник"
	go build -o $(SOCKET_BIN) $(SOCKET_FILENAMES) 

fclean:
	rm -rf $(SOCKET_BIN)

re: fclean all