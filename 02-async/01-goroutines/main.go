package main

import (
	"log"
	"math/rand"
	"time"
)

type User struct {
	Email string
}

type UserRepository interface {
	CreateUserAccount(u User) error
}

type NotificationsClient interface {
	SendNotification(u User) error
}

type NewsletterClient interface {
	AddToNewsletter(u User) error
}

type Handler struct {
	repository          UserRepository
	newsletterClient    NewsletterClient
	notificationsClient NotificationsClient
}

func NewHandler(
	repository UserRepository,
	newsletterClient NewsletterClient,
	notificationsClient NotificationsClient,
) Handler {
	return Handler{
		repository:          repository,
		newsletterClient:    newsletterClient,
		notificationsClient: notificationsClient,
	}
}

func (h Handler) SignUp(u User) error {
	if err := h.repository.CreateUserAccount(u); err != nil {
		return err
	}

	go func() {
		for {
			if err := h.newsletterClient.AddToNewsletter(u); err != nil {
				log.Println("failed to add user to newsletter, retrying...")
				// sleep random amount of time
				r := time.Duration(50+rand.Intn(100)) * time.Millisecond
				time.Sleep(r)
			} else {
				return
			}
		}
	}()
	go func() {
		for {
			if err := h.notificationsClient.SendNotification(u); err != nil {
				log.Println("failed to send notification, retrying...")
				r := time.Duration(50+rand.Intn(100)) * time.Millisecond
				time.Sleep(r)
			} else {
				return
			}
		}
	}()

	return nil
}
