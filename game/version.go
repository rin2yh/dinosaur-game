package game

// Version is the released version of this repository, rewritten by tagpr
// in the release pull request to match the git tag (see .tagpr); it is
// not edited by hand, and between releases it names the last release
// rather than the working tree.
//
// It exists for the firmware, since the board has neither a clock nor a
// version display: a frontend that carries this value is the only way to
// tell later which source a flashed board came from. Nothing reads it
// yet, so no binary carries the string until something does.
const Version = "0.0.0"
