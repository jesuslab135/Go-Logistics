package auth

import (
	"errors"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	TypeAccess  = "access"
	TypeRefresh = "refresh"
)

var ErrInvalidToken = errors.New("auth: invalid token")

// Claims are the JWT payload. Subject holds the employee id (string per the JWT
// spec); EmployeeID decodes it. CompanyID scopes every authenticated request to
// a tenant.
type Claims struct {
	CompanyID int64  `json:"company_id"`
	IsAdmin   bool   `json:"is_admin"`
	Type      string `json:"typ"`
	jwt.RegisteredClaims
}

func (c Claims) EmployeeID() int64 {
	id, _ := strconv.ParseInt(c.Subject, 10, 64)
	return id
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
}

type TokenService struct {
	secret     []byte
	issuer     string
	accessTTL  time.Duration
	refreshTTL time.Duration
	now        func() time.Time
}

func NewTokenService(secret, issuer string, accessTTL, refreshTTL time.Duration) *TokenService {
	return &TokenService{
		secret:     []byte(secret),
		issuer:     issuer,
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
		now:        time.Now,
	}
}

func (s *TokenService) Issue(employeeID, companyID int64, isAdmin bool) (TokenPair, error) {
	access, err := s.sign(employeeID, companyID, isAdmin, TypeAccess, s.accessTTL)
	if err != nil {
		return TokenPair{}, err
	}
	refresh, err := s.sign(employeeID, companyID, isAdmin, TypeRefresh, s.refreshTTL)
	if err != nil {
		return TokenPair{}, err
	}
	return TokenPair{
		AccessToken:  access,
		RefreshToken: refresh,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.accessTTL.Seconds()),
	}, nil
}

func (s *TokenService) sign(employeeID, companyID int64, isAdmin bool, typ string, ttl time.Duration) (string, error) {
	now := s.now()
	claims := Claims{
		CompanyID: companyID,
		IsAdmin:   isAdmin,
		Type:      typ,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatInt(employeeID, 10),
			Issuer:    s.issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.secret)
}

// Parse verifies signature, issuer, expiry, and the HS256 method, returning the
// claims. It does not check the token Type — callers decide access vs refresh.
func (s *TokenService) Parse(token string) (*Claims, error) {
	var claims Claims
	parsed, err := jwt.ParseWithClaims(token, &claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return s.secret, nil
	}, jwt.WithIssuer(s.issuer), jwt.WithValidMethods([]string{"HS256"}))
	if err != nil || !parsed.Valid {
		return nil, ErrInvalidToken
	}
	return &claims, nil
}

// ParseRefresh additionally requires the token to be a refresh token.
func (s *TokenService) ParseRefresh(token string) (*Claims, error) {
	claims, err := s.Parse(token)
	if err != nil {
		return nil, err
	}
	if claims.Type != TypeRefresh {
		return nil, ErrInvalidToken
	}
	return claims, nil
}
