package Feedback

import (
	"github.com/gofiber/fiber/v2"
	FB "github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/database/Feedback"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/models/Feedback"
)

// GetAllFeedbacksHandler returns all feedbacks
func GetAllFeedbacksHandler(c *fiber.Ctx) error {
	feedbacks, err := FB.GetAllFeedbacks()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(feedbacks)
}

// GetFeedbackByIDHandler returns a feedback by its ID
func GetFeedbackByIDHandler(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid id",
		})
	}
	if id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "id must be greater than 0",
		})
	}

	feedback, err := FB.GetFeedbackByID(id)
	if err != nil {
		if err.Error() == "feedback not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(feedback)
}

// CreateFeedbackHandler create a new feedback
func CreateFeedbackHandler(c *fiber.Ctx) error {
	feedback := new(Feedback.Feedback)
	if err := c.BodyParser(feedback); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	if feedback.Date.String() == "" || feedback.Channel == "" || feedback.Text == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "id, channel and text are required",
		})
	}

	createdFeedback, err := FB.CreateFeedback(*feedback)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.Status(fiber.StatusCreated).JSON(createdFeedback)
}

// UpdateFeedbackHandler updates an existing feedback
func UpdateFeedbackHandler(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid id",
		})
	}
	if id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "id must be greater than 0",
		})
	}

	feedback := new(Feedback.Feedback)
	if err = c.BodyParser(feedback); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	feedback.BaseModel.Id = id
	updatedFeedback, err := FB.UpdateFeedback(*feedback)
	if err != nil {
		if err.Error() == "feedback not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(updatedFeedback)
}

// DeleteFeedbackHandler deletes a feedback by its ID
func DeleteFeedbackHandler(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid id",
		})
	}
	if id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "id must be greater than 0",
		})
	}

	err = FB.DeleteFeedback(id)
	if err != nil {
		if err.Error() == "feedback not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// GetFeedbacksByChannelHandler returns feedbacks by channel
func GetFeedbacksByChannelHandler(c *fiber.Ctx) error {
	channel := c.Params("channel")
	if channel == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "channel is required",
		})
	}

	feedbacks, err := FB.GetFeedbacksByChannel(channel)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(feedbacks)
}
