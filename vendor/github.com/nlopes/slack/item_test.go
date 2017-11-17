package slack

import "testing"

func TestNewMessageItem(t *testing.T) {
	c := "C1"
	m := &Message{}
	mi := NewMessageItem(c, m)
	if mi.Type != TYPE_MESSAGE {
		t.Errorf("want Type %s, got %s", mi.Type, TYPE_MESSAGE)
	}
	if mi.Channel != c {
		t.Errorf("got Channel %s, want %s", mi.Channel, c)
	}
	if mi.Message != m {
		t.Errorf("got Message %v, want %v", mi.Message, m)
	}
}

func TestNewFileItem(t *testing.T) {
	f := &File{}
	fi := NewFileItem(f)
	if fi.Type != TYPE_FILE {
		t.Errorf("got Type %s, want %s", fi.Type, TYPE_FILE)
	}
	if fi.File != f {
		t.Errorf("got File %v, want %v", fi.File, f)
	}
}

func TestNewFileCommentItem(t *testing.T) {
	f := &File{}
	c := &Comment{}
	fci := NewFileCommentItem(f, c)
	if fci.Type != TYPE_FILE_COMMENT {
		t.Errorf("got Type %s, want %s", fci.Type, TYPE_FILE_COMMENT)
	}
	if fci.File != f {
		t.Errorf("got File %v, want %v", fci.File, f)
	}
	if fci.Comment != c {
		t.Errorf("got Comment %v, want %v", fci.Comment, c)
	}
}

func TestNewChannelItem(t *testing.T) {
	c := "C1"
	ci := NewChannelItem(c)
	if ci.Type != TYPE_CHANNEL {
		t.Errorf("got Type %s, want %s", ci.Type, TYPE_CHANNEL)
	}
	if ci.Channel != "C1" {
		t.Errorf("got Channel %v, want %v", ci.Channel, "C1")
	}
}

func TestNewIMItem(t *testing.T) {
	c := "D1"
	ci := NewIMItem(c)
	if ci.Type != TYPE_IM {
		t.Errorf("got Type %s, want %s", ci.Type, TYPE_IM)
	}
	if ci.Channel != "D1" {
		t.Errorf("got Channel %v, want %v", ci.Channel, "D1")
	}
}

func TestNewGroupItem(t *testing.T) {
	c := "G1"
	ci := NewGroupItem(c)
	if ci.Type != TYPE_GROUP {
		t.Errorf("got Type %s, want %s", ci.Type, TYPE_GROUP)
	}
	if ci.Channel != "G1" {
		t.Errorf("got Channel %v, want %v", ci.Channel, "G1")
	}
}

func TestNewRefToMessage(t *testing.T) {
	ref := NewRefToMessage("chan", "ts")
	if got, want := ref.Channel, "chan"; got != want {
		t.Errorf("Channel got %s, want %s", got, want)
	}
	if got, want := ref.Timestamp, "ts"; got != want {
		t.Errorf("Timestamp got %s, want %s", got, want)
	}
	if got, want := ref.File, ""; got != want {
		t.Errorf("File got %s, want %s", got, want)
	}
	if got, want := ref.Comment, ""; got != want {
		t.Errorf("Comment got %s, want %s", got, want)
	}
}

func TestNewRefToFile(t *testing.T) {
	ref := NewRefToFile("file")
	if got, want := ref.Channel, ""; got != want {
		t.Errorf("Channel got %s, want %s", got, want)
	}
	if got, want := ref.Timestamp, ""; got != want {
		t.Errorf("Timestamp got %s, want %s", got, want)
	}
	if got, want := ref.File, "file"; got != want {
		t.Errorf("File got %s, want %s", got, want)
	}
	if got, want := ref.Comment, ""; got != want {
		t.Errorf("Comment got %s, want %s", got, want)
	}
}

func TestNewRefToComment(t *testing.T) {
	ref := NewRefToComment("file_comment")
	if got, want := ref.Channel, ""; got != want {
		t.Errorf("Channel got %s, want %s", got, want)
	}
	if got, want := ref.Timestamp, ""; got != want {
		t.Errorf("Timestamp got %s, want %s", got, want)
	}
	if got, want := ref.File, ""; got != want {
		t.Errorf("File got %s, want %s", got, want)
	}
	if got, want := ref.Comment, "file_comment"; got != want {
		t.Errorf("Comment got %s, want %s", got, want)
	}
}
