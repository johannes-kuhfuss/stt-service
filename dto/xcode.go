package dto

type XcodeRequest struct {
	SourceFilePath string `json:"source_file_path" binding:"required" san:"trim,xss"`
}

type Xcode struct {
	XcodeDate      string
	SourceFileName string
	Status         string
	TargetFileName string
}
