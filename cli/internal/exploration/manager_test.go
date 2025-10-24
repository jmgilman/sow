package exploration

import (
	"testing"
)

func TestAddFile(t *testing.T) {
	ctx, cleanup := setupTestContext(t)
	defer cleanup()

	// Create exploration
	if err := InitExploration(ctx, "test", "explore/test"); err != nil {
		t.Fatalf("InitExploration() failed: %v", err)
	}

	// Add file
	path := "oauth-research.md"
	description := "OAuth 2.0 research notes"
	tags := []string{"oauth", "authentication"}

	if err := AddFile(ctx, path, description, tags); err != nil {
		t.Fatalf("AddFile() failed: %v", err)
	}

	// Verify file was added
	index, err := LoadIndex(ctx)
	if err != nil {
		t.Fatalf("LoadIndex() failed: %v", err)
	}

	if len(index.Files) != 1 {
		t.Fatalf("Files length = %d, want 1", len(index.Files))
	}

	file := index.Files[0]
	if file.Path != path {
		t.Errorf("File path = %q, want %q", file.Path, path)
	}

	if file.Description != description {
		t.Errorf("File description = %q, want %q", file.Description, description)
	}

	if len(file.Tags) != 2 {
		t.Fatalf("Tags length = %d, want 2", len(file.Tags))
	}

	// Test adding duplicate file
	if err := AddFile(ctx, path, "Different description", tags); err != ErrFileExists {
		t.Errorf("AddFile() duplicate = %v, want ErrFileExists", err)
	}
}

func TestUpdateFile(t *testing.T) {
	ctx, cleanup := setupTestContext(t)
	defer cleanup()

	// Create exploration and add file
	if err := InitExploration(ctx, "test", "explore/test"); err != nil {
		t.Fatalf("InitExploration() failed: %v", err)
	}

	path := "test.md"
	if err := AddFile(ctx, path, "Original description", []string{"tag1"}); err != nil {
		t.Fatalf("AddFile() failed: %v", err)
	}

	// Update file
	newDescription := "Updated description"
	newTags := []string{"tag1", "tag2", "tag3"}

	if err := UpdateFile(ctx, path, newDescription, newTags); err != nil {
		t.Fatalf("UpdateFile() failed: %v", err)
	}

	// Verify update
	index, err := LoadIndex(ctx)
	if err != nil {
		t.Fatalf("LoadIndex() failed: %v", err)
	}

	if len(index.Files) != 1 {
		t.Fatalf("Files length = %d, want 1", len(index.Files))
	}

	file := index.Files[0]
	if file.Description != newDescription {
		t.Errorf("Description after update = %q, want %q", file.Description, newDescription)
	}

	if len(file.Tags) != 3 {
		t.Errorf("Tags length after update = %d, want 3", len(file.Tags))
	}

	// Test updating non-existent file
	if err := UpdateFile(ctx, "nonexistent.md", "desc", []string{"tag"}); err != ErrFileNotFound {
		t.Errorf("UpdateFile() non-existent = %v, want ErrFileNotFound", err)
	}
}

func TestRemoveFile(t *testing.T) {
	ctx, cleanup := setupTestContext(t)
	defer cleanup()

	// Create exploration and add files
	if err := InitExploration(ctx, "test", "explore/test"); err != nil {
		t.Fatalf("InitExploration() failed: %v", err)
	}

	if err := AddFile(ctx, "file1.md", "File 1", []string{"tag1"}); err != nil {
		t.Fatalf("AddFile() file1 failed: %v", err)
	}

	if err := AddFile(ctx, "file2.md", "File 2", []string{"tag2"}); err != nil {
		t.Fatalf("AddFile() file2 failed: %v", err)
	}

	// Remove first file
	if err := RemoveFile(ctx, "file1.md"); err != nil {
		t.Fatalf("RemoveFile() failed: %v", err)
	}

	// Verify removal
	index, err := LoadIndex(ctx)
	if err != nil {
		t.Fatalf("LoadIndex() failed: %v", err)
	}

	if len(index.Files) != 1 {
		t.Fatalf("Files length after removal = %d, want 1", len(index.Files))
	}

	if index.Files[0].Path != "file2.md" {
		t.Errorf("Remaining file = %q, want %q", index.Files[0].Path, "file2.md")
	}

	// Test removing non-existent file
	if err := RemoveFile(ctx, "nonexistent.md"); err != ErrFileNotFound {
		t.Errorf("RemoveFile() non-existent = %v, want ErrFileNotFound", err)
	}
}

func TestGetFile(t *testing.T) {
	ctx, cleanup := setupTestContext(t)
	defer cleanup()

	// Create exploration and add file
	if err := InitExploration(ctx, "test", "explore/test"); err != nil {
		t.Fatalf("InitExploration() failed: %v", err)
	}

	path := "test.md"
	description := "Test file"
	tags := []string{"test", "example"}

	if err := AddFile(ctx, path, description, tags); err != nil {
		t.Fatalf("AddFile() failed: %v", err)
	}

	// Get file
	file, err := GetFile(ctx, path)
	if err != nil {
		t.Fatalf("GetFile() failed: %v", err)
	}

	if file.Path != path {
		t.Errorf("File path = %q, want %q", file.Path, path)
	}

	if file.Description != description {
		t.Errorf("File description = %q, want %q", file.Description, description)
	}

	// Test getting non-existent file
	_, err = GetFile(ctx, "nonexistent.md")
	if err != ErrFileNotFound {
		t.Errorf("GetFile() non-existent = %v, want ErrFileNotFound", err)
	}
}

func TestListFiles(t *testing.T) {
	ctx, cleanup := setupTestContext(t)
	defer cleanup()

	// Create exploration
	if err := InitExploration(ctx, "test", "explore/test"); err != nil {
		t.Fatalf("InitExploration() failed: %v", err)
	}

	// List with no files
	files, err := ListFiles(ctx)
	if err != nil {
		t.Fatalf("ListFiles() empty failed: %v", err)
	}

	if len(files) != 0 {
		t.Errorf("ListFiles() empty length = %d, want 0", len(files))
	}

	// Add files
	if err := AddFile(ctx, "file1.md", "File 1", []string{"tag1"}); err != nil {
		t.Fatalf("AddFile() file1 failed: %v", err)
	}

	if err := AddFile(ctx, "file2.md", "File 2", []string{"tag2"}); err != nil {
		t.Fatalf("AddFile() file2 failed: %v", err)
	}

	// List files
	files, err = ListFiles(ctx)
	if err != nil {
		t.Fatalf("ListFiles() failed: %v", err)
	}

	if len(files) != 2 {
		t.Errorf("ListFiles() length = %d, want 2", len(files))
	}
}

func TestUpdateStatus(t *testing.T) {
	ctx, cleanup := setupTestContext(t)
	defer cleanup()

	// Create exploration
	if err := InitExploration(ctx, "test", "explore/test"); err != nil {
		t.Fatalf("InitExploration() failed: %v", err)
	}

	// Update to completed
	if err := UpdateStatus(ctx, "completed"); err != nil {
		t.Fatalf("UpdateStatus() to completed failed: %v", err)
	}

	// Verify status
	index, err := LoadIndex(ctx)
	if err != nil {
		t.Fatalf("LoadIndex() failed: %v", err)
	}

	if index.Exploration.Status != "completed" {
		t.Errorf("Status = %q, want %q", index.Exploration.Status, "completed")
	}

	// Update to abandoned
	if err := UpdateStatus(ctx, "abandoned"); err != nil {
		t.Fatalf("UpdateStatus() to abandoned failed: %v", err)
	}

	// Verify status
	index, err = LoadIndex(ctx)
	if err != nil {
		t.Fatalf("LoadIndex() after abandoned failed: %v", err)
	}

	if index.Exploration.Status != "abandoned" {
		t.Errorf("Status = %q, want %q", index.Exploration.Status, "abandoned")
	}
}
