package Analysis

import (
	"github.com/gofiber/fiber/v2"
	AN "github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/database/Analysis"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/models/Analysis"
	"strconv"
	"time"
)

// GetAllAnalysesHandler returns all analyses
func GetAllAnalysesHandler(c *fiber.Ctx) error {
	analyses, err := AN.GetAllAnalyses()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(analyses)
}

// GetAnalysisByIDHandler returns an analysis by its ID
func GetAnalysisByIDHandler(c *fiber.Ctx) error {
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

	analysis, err := AN.GetAnalysisByID(id)
	if err != nil {
		if err.Error() == "analysis not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(analysis)
}

// AddAnalysisHandler creates a new analysis
func AddAnalysisHandler(c *fiber.Ctx) error {
	analysis := new(Analysis.Analysis)
	if err := c.BodyParser(analysis); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Validate required fields
	if analysis.SentimentScore < -1 || analysis.SentimentScore > 1 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "sentiment_score must be between -1 and 1",
		})
	}

	if analysis.Topic == "" || analysis.FeedbackID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "topic and feedback_id are required",
		})
	}

	// Set creation time
	analysis.CreatedAt = time.Now()

	createdAnalysis, err := AN.AddAnalysis(*analysis)
	if err != nil {
		if err.Error() == "feedback not found: the referenced feedback_id does not exist" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "feedback not found: the referenced feedback_id does not exist",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.Status(fiber.StatusCreated).JSON(createdAnalysis)
}

// UpdateAnalysisHandler updates an existing analysis
func UpdateAnalysisHandler(c *fiber.Ctx) error {
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

	analysis := new(Analysis.Analysis)
	if err = c.BodyParser(analysis); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Validate sentiment score if provided
	if analysis.SentimentScore < -1 || analysis.SentimentScore > 1 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "sentiment_score must be between -1 and 1",
		})
	}

	analysis.BaseModel.Id = id
	updatedAnalysis, err := AN.UpdateAnalysis(*analysis)
	if err != nil {
		if err.Error() == "analysis not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(updatedAnalysis)
}

// DeleteAnalysisHandler deletes an analysis by its ID
func DeleteAnalysisHandler(c *fiber.Ctx) error {
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

	err = AN.DeleteAnalysis(id)
	if err != nil {
		if err.Error() == "analysis not found" {
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

// GetAnalysisByFeedbackIDHandler returns an analysis for a specific feedback
func GetAnalysisByFeedbackIDHandler(c *fiber.Ctx) error {
	feedbackID := c.Params("feedback_id")
	if feedbackID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "feedback_id is required",
		})
	}

	analysis, err := AN.GetAnalysisByFeedbackID(feedbackID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(analysis)
}

// GetAnalysesByTopicHandler returns analyses filtered by topic
func GetAnalysesByTopicHandler(c *fiber.Ctx) error {
	topic := c.Params("topic")
	if topic == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "topic is required",
		})
	}

	analyses, err := AN.GetAnalysesByTopic(topic)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(analyses)
}

// GetAnalysesBySentimentRangeHandler returns analyses with sentiment in a specific range
func GetAnalysesBySentimentRangeHandler(c *fiber.Ctx) error {
	minScore, err := strconv.ParseFloat(c.Query("min", "-1"), 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid min score",
		})
	}

	maxScore, err := strconv.ParseFloat(c.Query("max", "1"), 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid max score",
		})
	}

	// Validate score range
	if minScore < -1 || minScore > 1 || maxScore < -1 || maxScore > 1 || minScore > maxScore {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid score range: min and max must be between -1 and 1, and min must be <= max",
		})
	}

	analyses, err := AN.GetAnalysesBySentimentRange(minScore, maxScore)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(analyses)
}
