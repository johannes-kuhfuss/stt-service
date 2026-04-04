package dto

type SttRequest struct {
	SourceFilePath string `json:"source_file_path" binding:"required" san:"trim,xss"`
}

type Stt struct {
	SttDate        string
	SourceFileName string
	Status         string
	TextFileName   string
}
