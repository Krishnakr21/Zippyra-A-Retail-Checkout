package main

import (
	"context"
	"encoding/json"
	"time"

	"github.com/zippyra/backend/shared/kafka"
	"github.com/zippyra/backend/shared/logger"
)

type EventConsumer struct {
	repo     Repository
	producer *kafka.Producer
}

func NewEventConsumer(repo Repository, producer *kafka.Producer) *EventConsumer {
	return &EventConsumer{
		repo:     repo,
		producer: producer,
	}
}

func (c *EventConsumer) ProcessOrderCompleted(ctx context.Context, value []byte) error {
	var payload OrderCompletedPayload
	if err := json.Unmarshal(value, &payload); err != nil {
		logger.Error("Failed to unmarshal order.completed event: %v", err)
		return err
	}

	if payload.OrderID == "" || payload.UserID == "" {
		logger.Warn("Received order.completed event with missing order_id or user_id")
		return nil
	}

	earnedPoints, oldTier, newTier, isUpgraded, newBalance, err := c.repo.EarnPointsTx(ctx, payload.OrderID, payload.UserID, payload.TotalPaise)
	if err != nil {
		logger.Error("Failed to process EarnPointsTx for order %s: %v", payload.OrderID, err)
		return err
	}

	now := time.Now()
	// 1. Publish loyalty.points_earned
	pointsEarnedMsg := PointsEarnedPayload{
		UserID:       payload.UserID,
		OrderID:      payload.OrderID,
		PointsEarned: earnedPoints,
		NewBalance:   newBalance,
		Timestamp:    now,
	}
	_ = c.producer.PublishEvent(ctx, TopicPointsEarned, payload.UserID, pointsEarnedMsg)

	// 2. Publish loyalty.tier_upgraded if upgraded
	if isUpgraded {
		tierUpgradedMsg := TierUpgradedPayload{
			UserID:    payload.UserID,
			OldTier:   oldTier,
			NewTier:   newTier,
			Timestamp: now,
		}
		_ = c.producer.PublishEvent(ctx, TopicTierUpgraded, payload.UserID, tierUpgradedMsg)
		logger.Info("User %s upgraded from tier %s to %s on order %s!", payload.UserID, oldTier, newTier, payload.OrderID)
	}

	// 3. Process referral reward if this is referred user's first completed order
	refUserID, refPts, refedPts, rewarded, refErr := c.repo.ProcessFirstOrderReferralReward(ctx, payload.UserID, payload.OrderID)
	if refErr != nil {
		logger.Error("Failed to process referral reward for user %s order %s: %v", payload.UserID, payload.OrderID, refErr)
	} else if rewarded {
		logger.Info("Referral reward granted for first order! Referrer %s +%d pts, Referred %s +%d pts", refUserID, refPts, payload.UserID, refedPts)
	}

	logger.Info("Processed order.completed for order %s user %s: earned %d points, new balance %d", payload.OrderID, payload.UserID, earnedPoints, newBalance)
	return nil
}

func (c *EventConsumer) ProcessOrderReturned(ctx context.Context, value []byte) error {
	var payload OrderReturnedPayload
	if err := json.Unmarshal(value, &payload); err != nil {
		logger.Error("Failed to unmarshal order.returned event: %v", err)
		return err
	}

	if payload.OrderID == "" || payload.UserID == "" {
		logger.Warn("Received order.returned event with missing order_id or user_id")
		return nil
	}

	reversedPoints, newBalance, err := c.repo.ReversePointsTx(
		ctx,
		payload.OrderID,
		payload.UserID,
		payload.ReturnID,
		payload.ReturnedAmountPaise,
		payload.OriginalTotalPaise,
	)
	if err != nil {
		logger.Error("Failed to process ReversePointsTx for order %s: %v", payload.OrderID, err)
		return err
	}

	logger.Info("Processed order.returned for order %s user %s: reversed %d points, new balance %d", payload.OrderID, payload.UserID, reversedPoints, newBalance)
	return nil
}
