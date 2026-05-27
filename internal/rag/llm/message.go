package llm

import (
	"encoding/json"
	"fmt"
)

// ContentPart is one OpenAI-style multimodal segment (text or image_url).
type ContentPart struct {
	Type     string        `json:"type"`
	Text     string        `json:"text,omitempty"`
	ImageURL *ImageURLPart `json:"image_url,omitempty"`
}

// ImageURLPart references inline base64 or remote image data.
type ImageURLPart struct {
	URL    string `json:"url"`
	Detail string `json:"detail,omitempty"`
}

// ImageAttachment is a decoded zone photo ready for the user turn.
type ImageAttachment struct {
	AttachmentID int64
	MimeType     string
	DataURL      string // data:image/jpeg;base64,...
}

// Message is a chat role + content. Content may be plain text (most turns) or
// multimodal parts when vision attachments are present (Phase 30 WS6).
type Message struct {
	Role    string
	Content string        // used when Parts is nil
	Parts   []ContentPart // when non-nil, marshals as OpenAI "content" array
}

// TextContent returns the human-readable text for logging/tests.
func (m Message) TextContent() string {
	if len(m.Parts) == 0 {
		return m.Content
	}
	var b string
	for _, p := range m.Parts {
		if p.Type == "text" && p.Text != "" {
			if b != "" {
				b += "\n"
			}
			b += p.Text
		}
	}
	return b
}

// MarshalJSON emits either a string content field or a parts array.
func (m Message) MarshalJSON() ([]byte, error) {
	if len(m.Parts) > 0 {
		return json.Marshal(struct {
			Role    string        `json:"role"`
			Content []ContentPart `json:"content"`
		}{Role: m.Role, Content: m.Parts})
	}
	return json.Marshal(struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}{Role: m.Role, Content: m.Content})
}

// UnmarshalJSON accepts string or array content from upstream responses.
func (m *Message) UnmarshalJSON(data []byte) error {
	var raw struct {
		Role    string          `json:"role"`
		Content json.RawMessage `json:"content"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	m.Role = raw.Role
	if len(raw.Content) == 0 {
		return nil
	}
	if raw.Content[0] == '"' {
		return json.Unmarshal(raw.Content, &m.Content)
	}
	if raw.Content[0] == '[' {
		return json.Unmarshal(raw.Content, &m.Parts)
	}
	return fmt.Errorf("unsupported message content json")
}

// UserMessageWithImages builds the final user turn for vision chat.
func UserMessageWithImages(text string, images []ImageAttachment) Message {
	if len(images) == 0 {
		return Message{Role: "user", Content: text}
	}
	parts := make([]ContentPart, 0, 1+len(images))
	if text != "" {
		parts = append(parts, ContentPart{Type: "text", Text: text})
	}
	for _, img := range images {
		parts = append(parts, ContentPart{
			Type: "image_url",
			ImageURL: &ImageURLPart{
				URL:    img.DataURL,
				Detail: "low",
			},
		})
	}
	return Message{Role: "user", Parts: parts}
}
