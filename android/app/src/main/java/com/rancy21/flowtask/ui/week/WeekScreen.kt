package com.rancy21.flowtask.ui.week

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Add
import androidx.compose.material.icons.filled.Check
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import com.rancy21.flowtask.data.entity.TaskEntity
import com.rancy21.flowtask.ui.theme.PriorityColors
import org.koin.androidx.compose.koinViewModel
import java.time.LocalDate
import java.time.format.DateTimeFormatter
import java.time.format.TextStyle
import java.util.Locale

@Composable
fun WeekScreen(
    viewModel: WeekViewModel = koinViewModel(),
    onOpenTaskEditor: (String?) -> Unit = {},
) {
    val uiState by viewModel.uiState.collectAsState()

    Scaffold(
        floatingActionButton = {
            FloatingActionButton(onClick = { onOpenTaskEditor("") }) {
                Icon(Icons.Filled.Add, contentDescription = "New task")
            }
        },
    ) { innerPadding ->
        Box(modifier = Modifier.fillMaxSize().padding(innerPadding)) {
            if (uiState.isLoading) {
                CircularProgressIndicator(modifier = Modifier.align(Alignment.Center))
            } else if (uiState.tasks.isEmpty()) {
                Text(
                    "No tasks this week",
                    modifier = Modifier.align(Alignment.Center),
                    style = MaterialTheme.typography.bodyLarge,
                    color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.5f),
                )
            } else {
                val today = LocalDate.now()
                val dayFormatter = DateTimeFormatter.ofPattern("EEE, MMM d")
                val grouped = viewModel.tasksByDay()

                LazyColumn(
                    contentPadding = PaddingValues(16.dp),
                    verticalArrangement = Arrangement.spacedBy(4.dp),
                ) {
                    grouped.forEach { (date, tasks) ->
                        item(key = date.toString()) {
                            val dayName = date.dayOfWeek.getDisplayName(TextStyle.FULL, Locale.getDefault())
                            val isToday = date == today
                            DayHeader(dayName = dayName, dateStr = date.format(dayFormatter), isToday = isToday)
                        }
                        items(tasks, key = { it.id }) { task ->
                            WeekTaskCard(
                                task = task,
                                onMarkDone = { viewModel.markDone(task) },
                                onDelete = { viewModel.delete(task) },
                                onClickEdit = { onOpenTaskEditor(task.id) },
                            )
                        }
                    }
                }
            }
        }
    }
}

@Composable
private fun DayHeader(dayName: String, dateStr: String, isToday: Boolean) {
    Text(
        text = if (isToday) "$dayName — $dateStr  ← Today" else "$dayName — $dateStr",
        style = MaterialTheme.typography.titleSmall.copy(
            fontWeight = if (isToday) FontWeight.Bold else FontWeight.Normal,
        ),
        color = if (isToday) MaterialTheme.colorScheme.primary
        else MaterialTheme.colorScheme.onSurface.copy(alpha = 0.7f),
        modifier = Modifier.padding(top = 12.dp, bottom = 4.dp),
    )
}

@Composable
private fun WeekTaskCard(
    task: TaskEntity,
    onMarkDone: () -> Unit,
    onDelete: () -> Unit,
    onClickEdit: () -> Unit,
) {
    val priorityColor = PriorityColors.colorFor(task.priority)

    Card(
        modifier = Modifier.fillMaxWidth(),
        onClick = onClickEdit,
    ) {
        Row(
            modifier = Modifier.fillMaxWidth().padding(12.dp),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            Surface(
                modifier = Modifier
                    .width(4.dp)
                    .height(40.dp),
                color = priorityColor,
                shape = MaterialTheme.shapes.small,
            ) {}

            Spacer(modifier = Modifier.width(12.dp))

            Column(modifier = Modifier.weight(1f)) {
                Text(
                    task.title,
                    style = MaterialTheme.typography.titleMedium,
                    maxLines = 1,
                    overflow = TextOverflow.Ellipsis,
                )
                Text(
                    task.priority,
                    style = MaterialTheme.typography.labelSmall,
                    color = priorityColor,
                )
            }

            IconButton(onClick = onMarkDone) {
                Icon(Icons.Filled.Check, contentDescription = "Mark done")
            }
        }
    }
}
