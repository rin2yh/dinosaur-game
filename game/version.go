package game

// Version is rewritten by tagpr in the release pull request to match the
// git tag (see .tagpr), so between releases it names the last release,
// not the working tree. It is here for the firmware — the board has no
// clock and no version display — but nothing reads it yet, so no binary
// carries the string.
const Version = "0.0.0"
