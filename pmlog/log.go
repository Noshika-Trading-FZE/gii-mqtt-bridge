/*
 * Copyright: Pixel Networks <support@pixel-networks.com> 
 */

package pmlog

import (
    "log"
)

func LogDebug(message ...interface{}) {
    log.Println("debug:", message)
    return
}

func LogError(message ...interface{}) {
    log.Println("error:", message)
    return
}

func LogWarning(message ...interface{}) {
    log.Println("warning:", message)
    return
}

func LogInfo(message ...interface{}) {
    log.Println("info:", message)
    return
}

func LogDetail(message ...interface{}) {
    log.Println("detail:", message)
    return
}


//EOF


