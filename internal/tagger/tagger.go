package tagger

import (
	"fmt"

	"github.com/bogem/id3v2/v2"
	"github.com/svk/wav2mp3/internal/config"
)

// Apply записывает ID3v2-теги в готовый MP3-файл.
func Apply(mp3Path string, tags config.ID3Tags) error {
	tag, err := id3v2.Open(mp3Path, id3v2.Options{Parse: false})
	if err != nil {
		return fmt.Errorf("не удалось открыть MP3 для тегирования: %w", err)
	}
	defer tag.Close()

	tag.SetVersion(4) // ID3v2.4

	if tags.Title != "" {
		tag.SetTitle(tags.Title)
	}
	if tags.Artist != "" {
		tag.SetArtist(tags.Artist)
	}
	if tags.Album != "" {
		tag.SetAlbum(tags.Album)
	}
	if tags.Year != "" {
		tag.SetYear(tags.Year)
	}
	if tags.Genre != "" {
		tag.SetGenre(tags.Genre)
	}
	if tags.Track != "" {
		tag.AddTextFrame(tag.CommonID("Track number/Position in set"), id3v2.EncodingUTF8, tags.Track)
	}
	if tags.Comment != "" {
		tag.AddCommentFrame(id3v2.CommentFrame{
			Encoding:    id3v2.EncodingUTF8,
			Language:    "eng",
			Description: "",
			Text:        tags.Comment,
		})
	}

	if tags.Cover != "" {
		cover, err := LoadCover(tags.Cover)
		if err != nil {
			return err
		}
		tag.AddAttachedPicture(id3v2.PictureFrame{
			Encoding:    id3v2.EncodingUTF8,
			MimeType:    cover.MIMEType,
			PictureType: id3v2.PTFrontCover,
			Description: "Cover",
			Picture:     cover.Data,
		})
	}

	if err := tag.Save(); err != nil {
		return fmt.Errorf("не удалось сохранить теги: %w", err)
	}
	return nil
}
