package main

import (
	"github.com/zippyra/backend/shared/sms"
)

type SmsSender = sms.SmsSender
type LogSmsSender = sms.LogSmsSender
type TwilioSmsSender = sms.TwilioSmsSender

func NewTwilioSmsSender() *TwilioSmsSender {
	return sms.NewTwilioSmsSender()
}
