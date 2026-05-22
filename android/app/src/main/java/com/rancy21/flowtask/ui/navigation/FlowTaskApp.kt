package com.rancy21.flowtask.ui.navigation

import androidx.compose.foundation.layout.padding
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontWeight
import com.rancy21.flowtask.ui.editor.TaskEditorScreen
import com.rancy21.flowtask.ui.inbox.InboxScreen
import com.rancy21.flowtask.ui.inboxeditor.InboxEditorScreen
import com.rancy21.flowtask.ui.notes.NotesScreen
import com.rancy21.flowtask.ui.today.TodayScreen
import com.rancy21.flowtask.ui.week.WeekScreen

@Composable
fun FlowTaskApp() {
    var currentScreen by remember { mutableStateOf<Screen>(Screen.Today) }
    var editingTaskId by remember { mutableStateOf<String?>(null) }
    var editingInboxId by remember { mutableStateOf<String?>(null) }

    // Editor overlays take priority
    if (editingTaskId != null || editingTaskId == "") {
        TaskEditorScreen(
            taskId = editingTaskId?.ifEmpty { null },
            onClose = { editingTaskId = null },
        )
        return
    }

    if (editingInboxId != null || editingInboxId == "") {
        InboxEditorScreen(
            itemId = editingInboxId?.ifEmpty { null },
            onClose = { editingInboxId = null },
        )
        return
    }

    Scaffold(
        bottomBar = {
            NavigationBar {
                listOf(Screen.Today, Screen.Week, Screen.Inbox, Screen.Notes).forEach { screen ->
                    NavigationBarItem(
                        icon = { Icon(screen.icon, contentDescription = screen.label) },
                        label = {
                            Text(
                                screen.label,
                                fontWeight = if (currentScreen == screen) FontWeight.Bold else FontWeight.Normal,
                            )
                        },
                        selected = currentScreen == screen,
                        onClick = { currentScreen = screen },
                    )
                }
            }
        },
    ) { innerPadding ->
        Surface(modifier = Modifier.padding(innerPadding)) {
            when (currentScreen) {
                Screen.Today -> TodayScreen(
                    onOpenTaskEditor = { taskId -> editingTaskId = taskId },
                )
                Screen.Week -> WeekScreen(
                    onOpenTaskEditor = { taskId -> editingTaskId = taskId },
                )
                Screen.Inbox -> InboxScreen(
                    onOpenInboxEditor = { itemId -> editingInboxId = itemId },
                )
                Screen.Notes -> NotesScreen()
            }
        }
    }
}
