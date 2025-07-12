package modules

import "github.com/divideprojects/Alita_Robot/alita/i18n"

// tr provides a default English translator for use in package-level initializations
// where a request-specific language code isn't available.
var tr = i18n.New("en")
