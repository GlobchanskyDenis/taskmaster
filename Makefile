BACK_BIN		=	back
BACK_DIR		=	cmd/back
BACK_FILES		=	main.go
BACK_FILENAMES	=	$(addprefix $(BACK_DIR)/,$(BACK_FILES))

all : $(BACK_BIN)

$(BACK_BIN) : $(BACK_FILENAMES)
	@echo "компилирую все в бинарник"
	go build -o $(BACK_BIN) $(BACK_FILENAMES) 

fclean:
	rm -rf $(BACK_BIN)

re: fclean all