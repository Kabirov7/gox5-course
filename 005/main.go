package main

import (
	"errors"
	"fmt"
)

var ErrEmptyMessage = errors.New("пустое сообщение")

type Notifier interface {
	Send(msg string) error
}

type BaseNotifier struct {
	Channel string
}

func (b BaseNotifier) Log() {
	fmt.Println("Sending notification via channel:", b.Channel)
}

type EmailNotifier struct {
	BaseNotifier
	Address string
}

func (e EmailNotifier) Send(msg string) error {
	// TODO: верните ErrEmptyMessage для пустого msg; иначе выведите канал, адрес и сообщение.
	if msg == "" {
		return ErrEmptyMessage
	}
	fmt.Println(e.Channel, e.Address, msg)
	return nil
}

var _ Notifier = EmailNotifier{}

type PushNotifier struct {
	BaseNotifier BaseNotifier
	DeviceID     uint64
}

func (p PushNotifier) Send(msg string) error {
	// TODO: сначала проверьте msg, затем DeviceID; иначе выведите канал, устройство и сообщение.
	if msg == "" {
		return ErrEmptyMessage
	}
	if p.DeviceID == 0 {
		return &SendError{
			Channel: p.BaseNotifier.Channel,
			Reason:  "не задан device id",
		}
	}
	fmt.Println(p.BaseNotifier.Channel, p.DeviceID, msg)
	return nil
}

var _ Notifier = PushNotifier{}

type SendError struct {
	Channel string
	Reason  string
}

func (e *SendError) Error() string {
	return fmt.Sprintf("канал %s: %s", e.Channel, e.Reason)
}

var _ error = (*SendError)(nil)

type BrokenNotifier struct{}

func (BrokenNotifier) Send(string) error {
	panic("драйвер канала повреждён")
}

type Channel struct {
	Name   string
	Sender Notifier
}

func SafeSend(notifier Notifier, msg string) (err error) {
	// TODO: до вызова Send зарегистрируйте defer с recover и заполните err при панике.
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic recovered: %v", r)
		}
	}()
	return notifier.Send(msg)
}

func Dispatch(all []Channel, msg string) []error {
	// TODO: вызовите каждый Sender через SafeSend, оберните ошибки через %w и соберите их.
	var failures []error

	for _, channel := range all {
		if err := SafeSend(channel.Sender, msg); err != nil {
			failures = append(failures, fmt.Errorf("канал %s: %w", channel.Name, err))
		}
	}

	return failures
}

func main() {
	email := EmailNotifier{
		BaseNotifier: BaseNotifier{Channel: "email"},
		Address:      "example@mail.ru",
	}
	push := PushNotifier{
		BaseNotifier: BaseNotifier{Channel: "push"},
		DeviceID:     123456,
	}

	channels := []Channel{
		{
			Name:   "email",
			Sender: email,
		},
		{
			Name:   "broken",
			Sender: BrokenNotifier{},
		},
		{
			Name:   "push",
			Sender: push,
		},
	}

	failures := Dispatch(channels, "Заказ готов")
	for _, err := range failures {
		fmt.Println("ошибка:", err)
	}
	fmt.Println("Всего ошибок:", len(failures))

	emptyChannels := []Channel{
		{Name: "email", Sender: email},
		{Name: "push", Sender: push},
	}
	for _, err := range Dispatch(emptyChannels, "") {
		if errors.Is(err, ErrEmptyMessage) {
			fmt.Println("нужно ввести текст")
		}
		if errors.Is(fmt.Errorf("рассылка: %w", err), ErrEmptyMessage) {
			fmt.Println("пустое сообщение найдено после обёртки")
		}
	}

	invalidPush := PushNotifier{
		BaseNotifier: BaseNotifier{Channel: "push"},
	}
	for _, err := range Dispatch([]Channel{{Name: "push", Sender: invalidPush}}, "Заказ готов") {
		var sendErr *SendError
		if errors.As(err, &sendErr) {
			fmt.Println("канал:", sendErr.Channel, "причина:", sendErr.Reason)
		}
	}
}
