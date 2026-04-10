package model_test

import (
	"strings"
	"testing"

	"github.com/robinovitch61/kl/internal/k8s/k8s_log"
	"github.com/robinovitch61/kl/internal/k8s/k8s_model"
	"github.com/robinovitch61/kl/internal/model"
	"github.com/robinovitch61/kl/internal/style"
	"github.com/robinovitch61/viewport/viewport/item"
)

func makePageLog(content string, timestamp string, name *k8s_model.ContainerNameAndPrefix, prettyPrinted bool, theme *style.Theme) model.PageLog {
	log := &k8s_log.Log{
		ContentItem: item.NewItem(content),
	}
	var containerNames *model.PageLogContainerNames
	if name != nil {
		containerNames = &model.PageLogContainerNames{
			Short: *name,
			Full:  *name,
		}
	}
	return model.PageLog{
		Log:              log,
		ContainerNames:   containerNames,
		CurrentName:      name,
		CurrentTimestamp: timestamp,
		Theme:            theme,
		PrettyPrinted:    prettyPrinted,
	}
}

func hasAnsi(s string) bool {
	return strings.Contains(s, "\x1b[")
}

func TestContentForFile_NoPrefix(t *testing.T) {
	pl := makePageLog("hello world", "", nil, false, nil)
	got := pl.ContentForFile()
	if got != "hello world" {
		t.Errorf("expected %q, got %q", "hello world", got)
	}
}

func TestContentForFile_TimestampNoAnsi(t *testing.T) {
	theme := style.DefaultTheme()
	pl := makePageLog("hello", "12:00:00", nil, false, &theme)
	got := pl.ContentForFile()
	if hasAnsi(got) {
		t.Errorf("ContentForFile should not contain ANSI codes, got %q", got)
	}
	if !strings.HasPrefix(got, "12:00:00 ") {
		t.Errorf("expected timestamp prefix, got %q", got)
	}
	if !strings.HasSuffix(got, "hello") {
		t.Errorf("expected content suffix, got %q", got)
	}
}

func TestContentForFile_ContainerNameNoAnsi(t *testing.T) {
	theme := style.DefaultTheme()
	name := &k8s_model.ContainerNameAndPrefix{Prefix: "my-pod", ContainerName: "web"}
	pl := makePageLog("hello", "", name, false, &theme)
	got := pl.ContentForFile()
	if hasAnsi(got) {
		t.Errorf("ContentForFile should not contain ANSI codes, got %q", got)
	}
	if !strings.Contains(got, "my-pod/web") {
		t.Errorf("expected container name in output, got %q", got)
	}
}

func TestContentForFile_TimestampAndContainerNameNoAnsi(t *testing.T) {
	theme := style.DefaultTheme()
	name := &k8s_model.ContainerNameAndPrefix{Prefix: "my-pod", ContainerName: "web"}
	pl := makePageLog("log line", "12:00:00", name, false, &theme)
	got := pl.ContentForFile()
	if hasAnsi(got) {
		t.Errorf("ContentForFile should not contain ANSI codes, got %q", got)
	}
	if !strings.HasPrefix(got, "12:00:00 my-pod/web log line") {
		t.Errorf("unexpected output %q", got)
	}
}

func TestGetItem_HasAnsiWithTheme(t *testing.T) {
	theme := style.DefaultTheme()
	name := &k8s_model.ContainerNameAndPrefix{Prefix: "my-pod", ContainerName: "web"}
	pl := makePageLog("hello", "12:00:00", name, false, &theme)
	got := pl.GetItem().Content()
	if !hasAnsi(got) {
		t.Errorf("GetItem().Content() should contain ANSI codes with a theme, got %q", got)
	}
}

func TestContentForFile_PrettyPrintedNoAnsi(t *testing.T) {
	theme := style.DefaultTheme()
	name := &k8s_model.ContainerNameAndPrefix{Prefix: "my-pod", ContainerName: "web"}
	pl := makePageLog(`{"key":"value","num":1}`, "12:00:00", name, true, &theme)
	got := pl.ContentForFile()
	if hasAnsi(got) {
		t.Errorf("ContentForFile should not contain ANSI codes for pretty-printed content, got %q", got)
	}
	if !strings.Contains(got, "12:00:00") {
		t.Errorf("expected timestamp in output, got %q", got)
	}
	if !strings.Contains(got, "my-pod/web") {
		t.Errorf("expected container name in output, got %q", got)
	}
	// pretty-printed JSON should be multi-line
	if !strings.Contains(got, "\n") {
		t.Errorf("expected multi-line pretty-printed output, got %q", got)
	}
}

func TestContentForFile_PrettyPrintedMatchesGetItemWithoutAnsi(t *testing.T) {
	theme := style.DefaultTheme()
	name := &k8s_model.ContainerNameAndPrefix{Prefix: "my-pod", ContainerName: "web"}
	pl := makePageLog(`{"key":"value"}`, "12:00:00", name, true, &theme)
	contentForFile := pl.ContentForFile()
	getItemNoAnsi := pl.GetItem().ContentNoAnsi()
	if contentForFile != getItemNoAnsi {
		t.Errorf("ContentForFile() and GetItem().ContentNoAnsi() should match\nContentForFile:    %q\nGetItem NoAnsi:    %q", contentForFile, getItemNoAnsi)
	}
}

func TestContentForFile_NonPrettyMatchesGetItemWithoutAnsi(t *testing.T) {
	theme := style.DefaultTheme()
	name := &k8s_model.ContainerNameAndPrefix{Prefix: "my-pod", ContainerName: "web"}
	pl := makePageLog("plain log line", "12:00:00", name, false, &theme)
	contentForFile := pl.ContentForFile()
	getItemNoAnsi := pl.GetItem().ContentNoAnsi()
	if contentForFile != getItemNoAnsi {
		t.Errorf("ContentForFile() and GetItem().ContentNoAnsi() should match\nContentForFile:    %q\nGetItem NoAnsi:    %q", contentForFile, getItemNoAnsi)
	}
}
