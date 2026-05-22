package com.rancy21.flowtask.ui.theme

import androidx.compose.ui.graphics.Color

// Blue + White + Black palette
val Primary = Color(0xFF3B82F6)
val OnPrimary = Color(0xFFFFFFFF)
val Secondary = Color(0xFF60A5FA)
val Tertiary = Color(0xFF93C5FD)
val Background = Color(0xFF0A0A0A)
val Surface = Color(0xFF1A1A2E)
val OnBackground = Color(0xFFF8FAFC)
val OnSurface = Color(0xFFE2E8F0)
val Error = Color(0xFFEF4444)

// Priority colors
val P1 = Color(0xFFF87171)
val P2 = Color(0xFFFBBF24)
val P3 = Color(0xFF34D399)

object PriorityColors {
    fun colorFor(priority: String): Color = when (priority) {
        "P1" -> P1
        "P2" -> P2
        "P3" -> P3
        else -> OnSurface
    }
}
