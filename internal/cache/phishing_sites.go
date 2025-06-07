package cache

import (
	"sync"
)

var PhishingSitesCache = &sync.Map{}
