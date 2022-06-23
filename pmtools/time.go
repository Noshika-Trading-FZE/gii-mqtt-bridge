/*
 * Copyright: Pixel Networks <support@pixel-networks.com> 
 */


package pmtools

import (
    "time"
)

func GetIsoTimestamp() string {
    return time.Now().Format(time.RFC3339)
}


//EOF

