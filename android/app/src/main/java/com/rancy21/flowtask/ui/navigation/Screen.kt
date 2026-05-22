package com.rancy21.flowtask.ui.navigation

import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.List
import androidx.compose.material.icons.filled.DateRange
import androidx.compose.material.icons.filled.Email
import androidx.compose.material.icons.filled.Home
import androidx.compose.ui.graphics.vector.ImageVector

sealed class Screen(
    val label: String,
    val icon: ImageVector,
) {
    data object Today : Screen("Today", Icons.Filled.Home)
    data object Week : Screen("Week", Icons.Filled.DateRange)
    data object Inbox : Screen("Inbox", Icons.Filled.Email)
    data object Notes : Screen("Notes", Icons.AutoMirrored.Filled.List)
}
