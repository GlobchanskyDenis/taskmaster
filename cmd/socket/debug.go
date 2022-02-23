package main

import (
	"github.com/GlobchanskyDenis/taskmaster.git/internal/supervisor"
	"time"
)

func debugSimple(newSupervisor supervisor.Supervisor) {
	time.Sleep(time.Second * 1)

	newSupervisor.StatusAll(Printer{})

	time.Sleep(time.Second * 1)

	println("status")
	if err := newSupervisor.Status("sample_simple_bin", Printer{}, 10); err != nil {
		println(err.Error())
	}

	time.Sleep(time.Second * 1)

	println("stop")
	if err := newSupervisor.Stop("sample_simple_bin", Printer{}); err != nil {
		println(err.Error())
	}

	time.Sleep(time.Second * 1)

	println("status")
	if err := newSupervisor.Status("sample_simple_bin", Printer{}, 10); err != nil {
		println(err.Error())
	}

	time.Sleep(time.Second * 3)

	println("start")
	if err := newSupervisor.Start("sample_simple_bin", Printer{}); err != nil {
		println(err.Error())
	}

	time.Sleep(time.Second * 1)

	println("status")
	if err := newSupervisor.Status("sample_simple_bin", Printer{}, 10); err != nil {
		println(err.Error())
	}

	time.Sleep(time.Second * 1)

	println("status")
	if err := newSupervisor.Status("sample_simple_bin", Printer{}, 10); err != nil {
		println(err.Error())
	}

	time.Sleep(time.Second * 1)

	println("status")
	if err := newSupervisor.Status("sample_simple_bin", Printer{}, 10); err != nil {
		println(err.Error())
	}

	time.Sleep(time.Second * 1)

	newSupervisor.StatusAll(Printer{})
}

func debugAutorestart(newSupervisor supervisor.Supervisor) {
	println("status")
	if err := newSupervisor.Status("sample_autorestart_bin", Printer{}, 10); err != nil {
		println(err.Error())
	}

	time.Sleep(time.Second * 10)

	println("status")
	if err := newSupervisor.Status("sample_autorestart_bin", Printer{}, 10); err != nil {
		println(err.Error())
	}
}

func debugLogs(newSupervisor supervisor.Supervisor) {
	time.Sleep(time.Second * 4)

	println("status")
	if err := newSupervisor.Status("sample_simple_bin", Printer{}, 10); err != nil {
		println(err.Error())
	}

	time.Sleep(time.Second * 10)

	println("status")
	if err := newSupervisor.Status("sample_simple_bin", Printer{}, 0); err != nil {
		println(err.Error())
	}
}