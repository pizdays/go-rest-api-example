package handler

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-rest-api-example/model"
	"github.com/go-rest-api-example/service"
	"github.com/go-rest-api-example/util"

	"github.com/gin-gonic/gin"
)

// Password represents password reset handler.
type Password struct {
	ps service.PasswordService
	es service.Email
}

const numberOfPasswordResetTokenByte = 128

func NewPasswordHandler(ps service.PasswordService, es service.Email) Password {
	return Password{ps, es}
}

// CreatePasswordReset creates password reset record or use existing one and
// sends password reset email.
//
// @Summary      Start password reset process
// @Tags         Password Reset
// @Description  Create password reset record or use existing one then send password reset email to user.
// @Param        email  query     string  true  "User email address for sending password reset email."
// @Success      200    {object}  model.SuccessResponse
// @Success      201    {object}  model.SuccessResponse
// @Failure      400    {object}  model.ErrorResponse  "email is required".
// @Failure      500    {object}  model.ErrorResponse
// @Router       /password/password-reset [POST]
func (h Password) CreatePasswordReset(c *gin.Context) {
	email := c.Query("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Message: "email is required",
		})

		return
	}

	ctx := c.MustGet("context").(context.Context)

	// User not in the system.
	if !h.ps.EmailExist(ctx, email) {
		c.JSON(http.StatusOK, model.SuccessResponse{
			Message: "success",
		})

		return
	}

	// Check whether password reset record with the email exist.
	prEmailExist := h.ps.CheckPasswordResetEmailExist(ctx, email)

	if prEmailExist {
		if h.ps.CheckPasswordResetEmailExpire(ctx, email) {
			h.createNewPasswordReset(c, email)
			return
		}

		// PasswordService reset record does not expire so use the existing record.
		pr, err := h.ps.GetPasswordReset(ctx, email)
		if err != nil {
			_ = c.Error(err)

			c.JSON(http.StatusInternalServerError, model.ErrorResponse{
				Message: err.Error(),
			})

			return
		}

		if err := h.es.SendPasswordResetEmail(pr); err != nil {
			_ = c.Error(err)

			c.JSON(http.StatusInternalServerError, model.ErrorResponse{
				Message: err.Error(),
			})

			return
		}
	}

	if !prEmailExist {
		h.createNewPasswordReset(c, email)
		return
	}

	c.JSON(http.StatusCreated, model.SuccessResponse{
		Message: "success",
	})
}

// createNewPasswordReset create new password reset record, format response if
// there is error, and return if there is an error.
func (h Password) createNewPasswordReset(c *gin.Context, email string) {
	ctx := c.MustGet("context").(context.Context)

	prOld, err := h.ps.GetPasswordReset(ctx, email)
	if err != nil {
		_ = c.Error(err)

		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Message: err.Error(),
		})

		return
	}

	if prOld.Email != "" {
		err = h.ps.DeletePasswordResetRecord(ctx, prOld.Email, prOld.Token)
		if err != nil {
			_ = c.Error(err)

			c.JSON(http.StatusInternalServerError, model.ErrorResponse{
				Message: err.Error(),
			})

			return
		}
	}

	prNew, err := h.createPasswordReset(ctx, email)
	if err != nil {
		_ = c.Error(err)

		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Message: err.Error(),
		})

		return
	}

	err = h.es.SendPasswordResetEmail(prNew)
	if err != nil {
		_ = c.Error(err)

		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Message: err.Error(),
		})

		return
	}

	c.JSON(http.StatusCreated, model.SuccessResponse{
		Message: "success",
	})
}

func (h Password) createPasswordReset(ctx context.Context, email string) (model.PasswordReset, error) {
	token, err := util.GenerateRandomString(numberOfPasswordResetTokenByte)
	if err != nil {
		return model.PasswordReset{}, fmt.Errorf("handler.PasswordService.createPasswordReset: %w", err)
	}

	pr, err := h.ps.CreatePasswordReset(ctx, email, token)
	if err != nil {
		return model.PasswordReset{}, fmt.Errorf("handler.PasswordService.createPasswordReset: %w", err)
	}

	return pr, nil
}

// ValidatePasswordReset checks whether password reset record with given email
// and password reset token exist or expired. The combination of email and
// password reset token are valid when password reset record with given email
// and password reset token exists and does not expired.
//
// @Summary      Validate email and password reset token
// @Tags         Password Reset
// @Description  The combination of  email and password reset token are valid when password reset record with given email and password reset **exists** and **does not expired**.
// @Param        email  query     string  true  "Email to be validated"
// @Param        token  query     string  true  "PasswordService reset token to be validated"
// @Success      200    {object}  model.PasswordResetValidityResponse
// @Failure      400    {object}  model.ErrorResponse
// @Failure      500    {object}  model.ErrorResponse
// @Router       /password/validate-password-reset [GET]
func (h Password) ValidatePasswordReset(c *gin.Context) {
	var params model.PasswordResetValidationParams

	if err := c.ShouldBindQuery(&params); err != nil {
		_ = c.Error(err)

		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Message: err.Error(),
		})

		return
	}

	ctx := c.MustGet("context").(context.Context)

	valid, err := h.ps.ValidateEmailAndToken(ctx, params.Email, params.Token)
	if err != nil {
		_ = c.Error(err)

		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Message: err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, model.PasswordResetValidityResponse{
		Valid: valid,
	})
}

// CheckPasswordResetExpire checks whether password reset record with a
// combination of email and password reset token expired.
//
// @Summary      Check password reset expiration
// @Tags         Password Reset
// @Description  Check whether password reset record with a combination of email and password reset token expired.
// @Param        email  query     string  true  "Email to be checked for expiration"
// @Param        token  query     string  true  "PasswordService reset token to be checked for expiration"
// @Success      200    {object}  model.PasswordResetExpirationResponse
// @Failure      400    {object}  model.ErrorResponse
// @Failure      500    {object}  model.ErrorResponse
// @Router       /password/check-password-reset-expire [GET]
func (h Password) CheckPasswordResetExpire(c *gin.Context) {
	var params model.PasswordResetValidationParams

	if err := c.ShouldBindQuery(&params); err != nil {
		_ = c.Error(err)

		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Message: err.Error(),
		})

		return
	}

	ctx := c.MustGet("context").(context.Context)

	expired, err := h.ps.CheckPasswordResetExpire(ctx,
		params.Email,
		params.Token)
	if err != nil {
		_ = c.Error(err)

		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Message: err.Error(),
		})

		return
	}

	if expired {
		// Delete password reset record
		err := h.ps.DeletePasswordResetRecord(ctx,
			params.Email,
			params.Token)
		if err != nil {
			_ = c.Error(err)

			c.JSON(http.StatusInternalServerError, model.ErrorResponse{
				Message: err.Error(),
			})

			return
		}
	}

	c.JSON(http.StatusOK, model.PasswordResetExpirationResponse{
		Expired: expired,
	})
}

// ResetPassword updates user current password with the given password.
//
// @Summary      Reset user password
// @Tags         Password Reset
// @Description  Update user current password with the given password.
// @Param        passwordResetParams  body  model.ResetPasswordParams  true  "PasswordService reset parameters"
// @Router       /password/password [PATCH]
func (h Password) ResetPassword(c *gin.Context) {
	var params model.ResetPasswordParams
	if err := c.ShouldBindJSON(&params); err != nil {
		_ = c.Error(err)

		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Message: err.Error(),
		})

		return
	}

	ctx := c.MustGet("context").(context.Context)

	exist, err := h.ps.ValidateEmailAndToken(ctx, params.Email, params.Token)
	if err != nil {
		_ = c.Error(err)

		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Message: err.Error(),
		})

		return
	}

	if !exist {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Message: "email and/or token is invalid",
		})

		return
	}

	if err := h.ps.ResetPassword(ctx, params); err != nil {
		_ = c.Error(err)

		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Message: err.Error(),
		})

		return
	}

	err = h.ps.DeletePasswordResetRecord(ctx, params.Email, params.Token)
	if err != nil {
		_ = c.Error(err)

		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Message: err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, model.SuccessResponse{
		Message: "success",
	})
}
