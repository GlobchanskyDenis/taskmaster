package supervisor

import (
	"github.com/GlobchanskyDenis/taskmaster.git/pkg/dto"
	"github.com/GlobchanskyDenis/taskmaster.git/pkg/constants"
	// "context"
	"fmt"
)

type unitMeta struct {
	pid      int
	name     string
	sender   chan<- dto.Command
	receiver <-chan dto.CommandResult

	statusCode uint
	status     string
	lastError  error
}

func newUnit(sender chan<- dto.Command, receiver <-chan dto.CommandResult) *unitMeta {
	return &unitMeta{
		sender: sender,
		receiver: receiver,
	}
}

func (u *unitMeta) getStatus() {
	fmt.Println("Отсылаю команду демону")
	u.sender <- dto.Command{
		Type: constants.COMMAND_STATUS,
	}
	fmt.Println("Отослал команду демону")
	result := <- u.receiver
	fmt.Println("Получил ответ от демона")

	u.pid = result.Pid
	u.name = result.Name
	u.statusCode = result.StatusCode
	u.status = result.Status
	u.lastError = result.Error
}

func (u *unitMeta) printShortStatus() {
	if u.statusCode == constants.STATUS_ACTIVE {
		fmt.Printf("[+] %5d %s\n", u.pid, u.name)
	} else {
		fmt.Printf("[-] %5d %s\n", u.pid, u.name)
	}
}
