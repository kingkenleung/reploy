package storage

import (
	"context"
	"fmt"
	"reploy/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	pool *pgxpool.Pool
}

func NewDB(databaseURL string) (*DB, error) {
	pool, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}
	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("database ping failed: %w", err)
	}
	return &DB{pool: pool}, nil
}

func (db *DB) Close() {
	db.pool.Close()
}

// --- User queries ---

func (db *DB) UpsertUser(ctx context.Context, u *models.User) (*models.User, error) {
	// lkh1 is always a teacher regardless of what is in the DB
	if u.Email == "lkh1@school.pyc.edu.hk" {
		u.Role = models.RoleTeacher
	}

	row := db.pool.QueryRow(ctx, `
		INSERT INTO users (google_id, email, pyccode, name, avatar_url, role)
		VALUES ($1, $2, $3, $4, $5, COALESCE(NULLIF($6, ''), 'student'))
		ON CONFLICT (google_id) DO UPDATE
		  SET name       = EXCLUDED.name,
		      avatar_url = EXCLUDED.avatar_url,
		      role       = CASE
		                     WHEN users.email = 'lkh1@school.pyc.edu.hk' THEN 'teacher'
		                     ELSE users.role
		                   END
		RETURNING id, google_id, email, pyccode, name, avatar_url, role, is_banned, created_at
	`, u.GoogleID, u.Email, u.PYCCode, u.Name, u.AvatarURL, string(u.Role))

	var out models.User
	err := row.Scan(&out.ID, &out.GoogleID, &out.Email, &out.PYCCode, &out.Name,
		&out.AvatarURL, &out.Role, &out.IsBanned, &out.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (db *DB) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	row := db.pool.QueryRow(ctx, `
		SELECT id, google_id, email, pyccode, name, avatar_url, role, is_banned, created_at
		FROM users WHERE id = $1
	`, id)
	var u models.User
	err := row.Scan(&u.ID, &u.GoogleID, &u.Email, &u.PYCCode, &u.Name,
		&u.AvatarURL, &u.Role, &u.IsBanned, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (db *DB) ListUsers(ctx context.Context) ([]*models.User, error) {
	rows, err := db.pool.Query(ctx, `
		SELECT id, google_id, email, pyccode, name, avatar_url, role, is_banned, created_at
		FROM users ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.GoogleID, &u.Email, &u.PYCCode, &u.Name,
			&u.AvatarURL, &u.Role, &u.IsBanned, &u.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, &u)
	}
	return users, nil
}

func (db *DB) SetUserBanned(ctx context.Context, id string, banned bool) error {
	_, err := db.pool.Exec(ctx, `UPDATE users SET is_banned = $1 WHERE id = $2`, banned, id)
	return err
}

func (db *DB) SetUserRole(ctx context.Context, id string, role models.Role) error {
	_, err := db.pool.Exec(ctx, `UPDATE users SET role = $1 WHERE id = $2`, string(role), id)
	return err
}

// --- App queries ---

func (db *DB) CreateApp(ctx context.Context, a *models.App) (*models.App, error) {
	cat := a.Category
	if cat == nil {
		cat = []byte("[]")
	}
	row := db.pool.QueryRow(ctx, `
		INSERT INTO apps (user_id, slug, title, description, html_content, category)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, user_id, slug, title, description, html_content, category, is_hidden, is_public, approved, created_at, updated_at
	`, a.UserID, a.Slug, a.Title, a.Description, a.HTMLContent, cat)

	var out models.App
	err := row.Scan(&out.ID, &out.UserID, &out.Slug, &out.Title, &out.Description,
		&out.HTMLContent, &out.Category, &out.IsHidden, &out.IsPublic, &out.Approved, &out.CreatedAt, &out.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (db *DB) GetAppByID(ctx context.Context, id string) (*models.App, error) {
	row := db.pool.QueryRow(ctx, `
		SELECT a.id, a.user_id, a.slug, a.title, a.description, a.html_content, a.category,
		       a.is_hidden, a.is_public, a.approved, a.created_at, a.updated_at, u.pyccode
		FROM apps a JOIN users u ON u.id = a.user_id
		WHERE a.id = $1
	`, id)
	var a models.App
	err := row.Scan(&a.ID, &a.UserID, &a.Slug, &a.Title, &a.Description,
		&a.HTMLContent, &a.Category, &a.IsHidden, &a.IsPublic, &a.Approved, &a.CreatedAt, &a.UpdatedAt, &a.OwnerPYCCode)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (db *DB) GetAppByPYCCodeAndSlug(ctx context.Context, pyccode, slug string) (*models.App, error) {
	row := db.pool.QueryRow(ctx, `
		SELECT a.id, a.user_id, a.slug, a.title, a.description, a.html_content, a.category,
		       a.is_hidden, a.is_public, a.approved, a.created_at, a.updated_at, u.pyccode
		FROM apps a JOIN users u ON u.id = a.user_id
		WHERE u.pyccode = $1 AND a.slug = $2
	`, pyccode, slug)
	var a models.App
	err := row.Scan(&a.ID, &a.UserID, &a.Slug, &a.Title, &a.Description,
		&a.HTMLContent, &a.Category, &a.IsHidden, &a.IsPublic, &a.Approved, &a.CreatedAt, &a.UpdatedAt, &a.OwnerPYCCode)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (db *DB) ListAppsByUser(ctx context.Context, userID string) ([]*models.App, error) {
	rows, err := db.pool.Query(ctx, `
		SELECT a.id, a.user_id, a.slug, a.title, a.description, a.html_content, a.category,
		       a.is_hidden, a.is_public, a.approved, a.created_at, a.updated_at, u.pyccode
		FROM apps a JOIN users u ON u.id = a.user_id
		WHERE a.user_id = $1 ORDER BY a.updated_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanApps(rows)
}

func (db *DB) ListAllApps(ctx context.Context) ([]*models.App, error) {
	rows, err := db.pool.Query(ctx, `
		SELECT a.id, a.user_id, a.slug, a.title, a.description, a.html_content, a.category,
		       a.is_hidden, a.is_public, a.approved, a.created_at, a.updated_at, u.pyccode
		FROM apps a JOIN users u ON u.id = a.user_id
		ORDER BY a.updated_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanApps(rows)
}

func (db *DB) ListApprovedApps(ctx context.Context) ([]*models.App, error) {
	rows, err := db.pool.Query(ctx, `
		SELECT a.id, a.user_id, a.slug, a.title, a.description, a.html_content, a.category,
		       a.is_hidden, a.is_public, a.approved, a.created_at, a.updated_at, u.pyccode
		FROM apps a JOIN users u ON u.id = a.user_id
		WHERE a.approved = true AND a.is_hidden = false AND u.is_banned = false
		ORDER BY a.updated_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanApps(rows)
}

func (db *DB) UpdateAppContent(ctx context.Context, id, title, description, htmlContent string, category []byte) error {
	if category == nil {
		category = []byte("[]")
	}
	_, err := db.pool.Exec(ctx, `
		UPDATE apps SET title=$1, description=$2, html_content=$3, category=$4, updated_at=NOW()
		WHERE id=$5
	`, title, description, htmlContent, category, id)
	return err
}

func (db *DB) SetAppHidden(ctx context.Context, id string, hidden bool) error {
	_, err := db.pool.Exec(ctx, `UPDATE apps SET is_hidden=$1 WHERE id=$2`, hidden, id)
	return err
}

func (db *DB) SetAppApproved(ctx context.Context, id string, approved bool) error {
	_, err := db.pool.Exec(ctx, `UPDATE apps SET approved=$1 WHERE id=$2`, approved, id)
	return err
}

func (db *DB) DeleteApp(ctx context.Context, id string) error {
	_, err := db.pool.Exec(ctx, `DELETE FROM apps WHERE id=$1`, id)
	return err
}

func (db *DB) GetAllCategories(ctx context.Context) ([]string, error) {
	rows, err := db.pool.Query(ctx, `SELECT DISTINCT jsonb_array_elements_text(category) AS cat FROM apps WHERE category IS NOT NULL AND category != '[]'::jsonb ORDER BY cat`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var cats []string
	for rows.Next() {
		var cat string
		if err := rows.Scan(&cat); err != nil {
			return nil, err
		}
		cats = append(cats, cat)
	}
	return cats, nil
}

func scanApps(rows interface {
	Next() bool
	Scan(...any) error
}) ([]*models.App, error) {
	var apps []*models.App
	for rows.Next() {
		var a models.App
		if err := rows.Scan(&a.ID, &a.UserID, &a.Slug, &a.Title, &a.Description,
			&a.HTMLContent, &a.Category, &a.IsHidden, &a.IsPublic, &a.Approved, &a.CreatedAt, &a.UpdatedAt, &a.OwnerPYCCode); err != nil {
			return nil, err
		}
		apps = append(apps, &a)
	}
	return apps, nil
}
