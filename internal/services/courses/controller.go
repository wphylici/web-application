package courses

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/goldlilya1612/diploma-backend/internal/models"
	"github.com/goldlilya1612/diploma-backend/internal/services/media"
	serv "github.com/goldlilya1612/diploma-backend/internal/transport/http"
	"github.com/goldlilya1612/diploma-backend/internal/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"net/http"
	"strings"
	"time"
)

type CoursesController struct {
	DB *gorm.DB
}

func NewCoursesController(DB *gorm.DB) *CoursesController {
	return &CoursesController{
		DB: DB,
	}
}

func (cc *CoursesController) CreateCourse(ctx *gin.Context) {
	var payload *models.CreateCourse

	imageUrl, err := media.FileUpload(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.HTTPResponse{
			Status:     serv.ErrResponseStatus,
			StatusCode: http.StatusInternalServerError,
			Message:    err.Error(),
		})
		return
	}

	err = ctx.ShouldBind(&payload)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, models.HTTPResponse{
			Status:     serv.ErrResponseStatus,
			StatusCode: http.StatusBadRequest,
			Message:    err.Error(),
		})
		return
	}

	currentUser := ctx.MustGet("currentUser").(models.User)
	if uuid.Must(uuid.Parse(payload.CreatorID)) != currentUser.ID {
		message := "Access denied"
		ctx.JSON(http.StatusForbidden, models.HTTPResponse{
			Status:     serv.ErrResponseStatus,
			StatusCode: http.StatusForbidden,
			Message:    message,
		})
		return
	}

	if payload.CreatorName != currentUser.Name {
		message := "The creator's name does not match the current user's name"
		ctx.JSON(http.StatusBadRequest, models.HTTPResponse{
			Status:     serv.ErrResponseStatus,
			StatusCode: http.StatusBadRequest,
			Message:    message,
		})
		return
	}

	now := time.Time{}
	newCourse := models.Course{
		Name:        payload.Name,
		CreatorName: payload.CreatorName,
		CreatorID:   uuid.Must(uuid.Parse(payload.CreatorID)),
		ImageUrl:    imageUrl,
		Category:    payload.Category,
		Description: payload.Description,

		CreatedAt: now,
		UpdatedAt: now,
	}

	res := cc.DB.Create(&newCourse)
	if res.Error != nil {
		ctx.JSON(http.StatusConflict, models.HTTPResponse{
			Status:     serv.ErrResponseStatus,
			StatusCode: http.StatusConflict,
			Message:    res.Error.Error(),
		})
		return
	}

	courseResponse := &models.CourseResponse{
		ID:          newCourse.ID,
		Name:        payload.Name,
		CreatorID:   uuid.Must(uuid.Parse(payload.CreatorID)),
		CreatorName: payload.CreatorName,
		ImageUrl:    imageUrl,
		Category:    payload.Category,
		Description: payload.Description,
		Route:       utils.Latinizer(payload.Name),

		CreatedAt: newCourse.CreatedAt,
		UpdatedAt: newCourse.UpdatedAt,
	}

	ctx.JSON(http.StatusCreated, models.HTTPResponse{
		Status:     serv.SuccessResponseStatus,
		StatusCode: http.StatusCreated,
		Data:       map[string]interface{}{"course": courseResponse},
	})
}

func (cc *CoursesController) GetCourse(ctx *gin.Context) {

	var coursesResponse []models.CourseResponse

	params := ctx.Request.URL.Query()
	ids := params["id"]

	for _, id := range ids {
		course := models.Course{}
		res := cc.DB.First(&course, "id = ?", fmt.Sprint(id))
		if res.Error != nil && strings.Contains(res.Error.Error(), "record not found") {
			message := fmt.Sprintf("Course with id=%s not found", id)
			ctx.JSON(http.StatusBadRequest, models.HTTPResponse{
				Status:     serv.ErrResponseStatus,
				StatusCode: http.StatusBadRequest,
				Message:    message,
			})
			return
		} else if res.Error != nil {
			ctx.JSON(http.StatusBadRequest, models.HTTPResponse{
				Status:     serv.ErrResponseStatus,
				StatusCode: http.StatusBadRequest,
				Message:    res.Error.Error(),
			})
			return
		}

		courseResponse := models.CourseResponse{
			ID:          course.ID,
			Name:        course.Name,
			CreatorID:   course.CreatorID,
			CreatorName: course.Name,
			//Image:       course.Image,
			Category:    course.Category,
			Description: course.Description,
			Route:       utils.Latinizer(course.Name),

			CreatedAt: course.CreatedAt,
			UpdatedAt: course.UpdatedAt,
		}
		coursesResponse = append(coursesResponse, courseResponse)
	}

	ctx.JSON(http.StatusOK, models.HTTPResponse{
		Status:     serv.SuccessResponseStatus,
		StatusCode: http.StatusOK,
		Data:       map[string]interface{}{"courses": coursesResponse},
	})
}

func (cc *CoursesController) GetCourses(ctx *gin.Context) {

	var coursesResponse []models.CourseResponse

	var courses []models.Course
	res := cc.DB.Find(&courses)
	if res.Error != nil {
		ctx.JSON(http.StatusBadRequest, models.HTTPResponse{
			Status:     serv.ErrResponseStatus,
			StatusCode: http.StatusBadRequest,
			Message:    res.Error.Error(),
		})
		return
	}

	for _, c := range courses {
		courseResponse := models.CourseResponse{
			ID:          c.ID,
			Name:        c.Name,
			CreatorID:   c.CreatorID,
			CreatorName: c.CreatorName,
			//Image:       c.Image,
			Category:    c.Category,
			Description: c.Description,
			Route:       utils.Latinizer(c.Name),

			CreatedAt: c.CreatedAt,
			UpdatedAt: c.UpdatedAt,
		}
		coursesResponse = append(coursesResponse, courseResponse)
	}

	ctx.JSON(http.StatusOK, models.HTTPResponse{
		Status:     serv.SuccessResponseStatus,
		StatusCode: http.StatusOK,
		Data:       map[string]interface{}{"courses": coursesResponse},
	})
}

func (cc *CoursesController) DeleteCourse(ctx *gin.Context) {

	var coursesResponse []models.CourseResponse

	params := ctx.Request.URL.Query()
	ids := params["id"]

	for _, id := range ids {

		course := models.Course{}
		res := cc.DB.Clauses(clause.Returning{}).Where("id = ?", fmt.Sprint(id)).Delete(&course)
		if res.Error != nil {
			ctx.JSON(http.StatusBadRequest, models.HTTPResponse{
				Status:     serv.ErrResponseStatus,
				StatusCode: http.StatusBadRequest,
				Message:    res.Error.Error(),
			})
			return
		}

		courseResponse := models.CourseResponse{
			ID:          course.ID,
			Name:        course.Name,
			CreatorID:   course.CreatorID,
			CreatorName: course.CreatorName,
			//Image:       course.Image,
			Category:    course.Category,
			Description: course.Description,
			Route:       utils.Latinizer(course.Name),

			CreatedAt: course.CreatedAt,
			UpdatedAt: course.UpdatedAt,
		}
		coursesResponse = append(coursesResponse, courseResponse)
	}

	ctx.JSON(http.StatusOK, models.HTTPResponse{
		Status:     serv.SuccessResponseStatus,
		StatusCode: http.StatusOK,
		Data:       map[string]interface{}{"deletedCourses": coursesResponse},
	})
}
