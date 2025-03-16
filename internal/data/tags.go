package data

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type Tag struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type ItemTag struct {
	ID     int32 `json:"id"`
	ItemID int32 `json:"item_id"`
	TagID  int32 `json:"tag_id"`
}

type TagModel struct {
	DB *sql.DB
}

var (
	ErrDuplicateName       = errors.New("duplicate name")
	ErrDuplicateItemTag    = errors.New("duplicate item tag")
	ErrItemIdDoesNotExists = errors.New("item id does not exist")
	ErrTagIdDoesNotExists  = errors.New("tag id does not exist")
)

func (m TagModel) InsertTag(tag *Tag) error {
	query := `
		INSERT INTO tags (name)
		VALUES ($1)
		RETURNING id`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, tag.Name).Scan(&tag.ID)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "tags_name_key"`:
			return ErrDuplicateName
		}
	}
	return err
}

func (m TagModel) DeleteTag(tagId int) error {
	query := `
		DELETE FROM tags
		WHERE id = $1
		`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	res, err := m.DB.ExecContext(ctx, query, tagId)
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	} else if affected == 0 {
		return ErrNoRecord
	}
	return err
}

func (m TagModel) GetTags() ([]*Tag, error) {
	query := `
	SELECT id, name
	FROM tags`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNoRecord
		default:
			return nil, err
		}
	}
	defer rows.Close()

	var tags []*Tag

	for rows.Next() {
		var tag Tag
		err := rows.Scan(
			&tag.ID,
			&tag.Name,
		)

		if err != nil {
			return nil, err
		}
		tags = append(tags, &tag)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return tags, nil
}

func (m TagModel) InsertItemTag(itemTag *ItemTag) error {
	query := `
	INSERT INTO item_tags (item_id, tag_id)
	VALUES ($1, $2)	
	RETURNING id`

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, itemTag.ItemID, itemTag.TagID).Scan(&itemTag.ID)
	if err != nil {
		switch {
		case err.Error() == `pq: insert or update on table "item_tags" violates foreign key constraint "item_tags_item_id_fkey"`:
			return ErrItemIdDoesNotExists
		case err.Error() == `pq: insert or update on table "item_tags" violates foreign key constraint "item_tags_tag_id_fkey"`:
			return ErrTagIdDoesNotExists
		case err.Error() == `pq: duplicate key value violates unique constraint "item_tags_unique"`:
			return ErrDuplicateItemTag
		}
	}
	return err
}

func (m TagModel) RemoveItemTag(itemId int, tagId int) error {
	query := `
	DELETE FROM item_tags WHERE item_id = $1 AND tag_id = $2`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	res, err := m.DB.ExecContext(ctx, query, itemId, tagId)
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	} else if affected == 0 {
		return ErrNoRecord
	}
	return err
}

func (m TagModel) GetTagsForItem(itemId int64) ([]string, error) {
	query := `
		SELECT name FROM tags
		INNER JOIN item_tags on tags.id = item_tags.tag_id and item_tags.item_id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, itemId)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNoRecord
		}
		return nil, err
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var tag string
		err := rows.Scan(&tag)
		if err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return tags, nil
}
