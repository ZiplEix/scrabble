package services

import (
	"github.com/ZiplEix/scrabble/api/models/response"
	"github.com/ZiplEix/scrabble/api/stats"
)

func GetAdminStats() (*response.AdminStatsResponse, error) {
	userCount, userPctChange, err := stats.GetActiveUsersCountAndVariance()
	if err != nil {
		return nil, err
	}

	gameCount, gamePctChange, err := stats.GetCreatedGamesCountAndVariance()
	if err != nil {
		return nil, err
	}

	activeGameCount, activeGamePctChange, err := stats.GetActiveGamesCountAndVariance()
	if err != nil {
		return nil, err
	}

	messageCount, messagePctChange, err := stats.GetMessageSendCountAndVariance()
	if err != nil {
		return nil, err
	}

	ticketCount, ticketPctChange, err := stats.GetTicketsCountAndVariance()
	if err != nil {
		return nil, err
	}

	return &response.AdminStatsResponse{
		ActiveUsersCount:        userCount,
		ActiveUsersPctChange:    userPctChange,
		CreatedGamesCount:       gameCount,
		CreatedGamesPctChange:   gamePctChange,
		ActiveGamesCount:        activeGameCount,
		ActiveGamesPctChange:    activeGamePctChange,
		SendedMessagesCount:     messageCount,
		SendedMessagesPctChange: messagePctChange,
		TicketsCreatedCount:     ticketCount,
		TicketsCreatedPctChange: ticketPctChange,
	}, nil
}
