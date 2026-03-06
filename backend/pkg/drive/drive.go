package drive

import (
	"context"
	"fmt"
	"io"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
)

type Service struct {
	svc      *drive.Service
	folderID string
}

func New(ctx context.Context, credentialsJSON []byte, folderID string) (*Service, error) {
	svc, err := drive.NewService(ctx,
		option.WithCredentialsJSON(credentialsJSON),
		option.WithScopes(drive.DriveFileScope),
	)
	if err != nil {
		return nil, fmt.Errorf("create drive service: %w", err)
	}
	return &Service{svc: svc, folderID: folderID}, nil
}

// Upload streams r to Google Drive, sets public read permission, and returns the file ID.
func (s *Service) Upload(ctx context.Context, name, contentType string, r io.Reader) (string, error) {
	f := &drive.File{
		Name:    name,
		Parents: []string{s.folderID},
	}
	res, err := s.svc.Files.Create(f).
		Media(r, googleapi.ContentType(contentType)).
		Context(ctx).
		Do()
	if err != nil {
		return "", fmt.Errorf("upload file: %w", err)
	}

	// Make publicly readable (anyone with link)
	_, err = s.svc.Permissions.Create(res.Id, &drive.Permission{
		Type: "anyone",
		Role: "reader",
	}).Context(ctx).Do()
	if err != nil {
		// Non-fatal: file is uploaded but not public; log externally if needed
		_ = s.svc.Files.Delete(res.Id).Context(ctx).Do()
		return "", fmt.Errorf("set file permission: %w", err)
	}

	return res.Id, nil
}

// Delete removes a file from Google Drive by its file ID.
func (s *Service) Delete(ctx context.Context, fileID string) error {
	if err := s.svc.Files.Delete(fileID).Context(ctx).Do(); err != nil {
		return fmt.Errorf("delete file %s: %w", fileID, err)
	}
	return nil
}
