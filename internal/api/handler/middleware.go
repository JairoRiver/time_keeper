package handler

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/JairoRiver/time_keeper/internal/util"
	"github.com/JairoRiver/time_keeper/pkg/token"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

const (
	authorizationHeaderKey  = "authorization"
	authorizationTypeBearer = "bearer"
	authorizationPayloadKey = "authorization_payload"
)

type UserInfo struct {
	UserId uuid.UUID
	Role   string
}

// AuthMiddleware creates a echo middleware for authorization
func (h *Handler) AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		authorizationHeader := ctx.Request().Header.Get(authorizationHeaderKey)

		if len(authorizationHeader) == 0 {
			return ctx.JSON(http.StatusUnauthorized, errorResponse("authorization header is not provided"))
		}

		fields := strings.Fields(authorizationHeader)
		if len(fields) < 2 {
			return ctx.JSON(http.StatusUnauthorized, errorResponse("invalid authorization header format"))
		}

		authorizationType := strings.ToLower(fields[0])
		if authorizationType != authorizationTypeBearer {
			return ctx.JSON(http.StatusUnauthorized, errorResponse("unsupported authorization type"))
		}

		accessToken := fields[1]

		//get userId from token (a malformed token is a client error, not a 500)
		userId, err := getUserIdFromToken(accessToken)
		if err != nil {
			h.log.Warn().Err(err).Msg("authMiddleware: parse token failed")
			return ctx.JSON(http.StatusUnauthorized, errorResponse("invalid or expired token"))
		}

		//Verify token
		payload, err := auxVerifyToken(h, userId, accessToken)
		if err != nil {
			h.log.Warn().Err(err).Msg("authMiddleware: verify token failed")
			return ctx.JSON(http.StatusUnauthorized, errorResponse("invalid or expired token"))
		}

		userInfo := UserInfo{
			UserId: payload.UserId,
			Role:   payload.Role,
		}
		ctx.Set(authorizationPayloadKey, userInfo)
		return next(ctx)
	}
}

// CookieMiddleware creates a echo middleware for extract values from cookies
type CookieInfo struct {
	Name  string
	Value string
}

func (h *Handler) CookieMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		cookie, err := ctx.Cookie(util.RefreshTokenName)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, errorResponse("authentication required"))
		}

		refreshTokenInfo := CookieInfo{
			Name:  cookie.Name,
			Value: cookie.Value,
		}

		//get userId from token (a malformed token is a client error, not a 500)
		userId, err := getUserIdFromToken(refreshTokenInfo.Value)
		if err != nil {
			h.log.Warn().Err(err).Msg("cookieMiddleware: parse token failed")
			return ctx.JSON(http.StatusUnauthorized, errorResponse("invalid or expired token"))
		}

		//Verify token
		payload, err := auxVerifyToken(h, userId, refreshTokenInfo.Value)
		if err != nil {
			h.log.Warn().Err(err).Msg("cookieMiddleware: verify token failed")
			return ctx.JSON(http.StatusUnauthorized, errorResponse("invalid or expired token"))
		}

		userInfo := UserInfo{
			UserId: payload.UserId,
			Role:   payload.Role,
		}

		ctx.Set(refreshTokenInfo.Name, userInfo)
		return next(ctx)
	}
}

// PageAuthMiddleware protects SSR page routes. Unlike CookieMiddleware it
// redirects to "/" instead of returning a 401 JSON response.
func (h *Handler) PageAuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		cookie, err := ctx.Cookie(util.RefreshTokenName)
		if err != nil {
			return ctx.Redirect(http.StatusSeeOther, "/")
		}
		userId, err := getUserIdFromToken(cookie.Value)
		if err != nil {
			return ctx.Redirect(http.StatusSeeOther, "/")
		}
		payload, err := auxVerifyToken(h, userId, cookie.Value)
		if err != nil {
			return ctx.Redirect(http.StatusSeeOther, "/")
		}
		ctx.Set(util.RefreshTokenName, UserInfo{UserId: payload.UserId, Role: payload.Role})
		return next(ctx)
	}
}

func getUserIdFromToken(tokenValue string) (uuid.UUID, error) {
	//get userId from token
	var userId uuid.UUID
	auxPayload := token.Payload{}
	auxToken, _, err := new(jwt.Parser).ParseUnverified(tokenValue, &auxPayload)
	if err != nil {
		return userId, err
	}

	// Extract UserId
	if claims, ok := auxToken.Claims.(*token.Payload); ok {
		userId = claims.UserId
	} else {
		return userId, fmt.Errorf("authMiddleware error extracting userId, error: %w", err)
	}

	return userId, nil
}

func auxVerifyToken(h *Handler, userId uuid.UUID, tokenValue string) (*token.Payload, error) {
	//Implement a new instance of tokenMaker
	secretKey, err := h.ctrl.GetUserSecretKey(context.Background(), userId)
	if err != nil {
		return nil, err
	}

	tokenMaker, err := token.NewJWTMaker(secretKey.SecretKey)
	if err != nil {
		return nil, err
	}

	payload, err := tokenMaker.VerifyToken(tokenValue)
	if err != nil {
		return nil, err
	}

	return payload, err
}
