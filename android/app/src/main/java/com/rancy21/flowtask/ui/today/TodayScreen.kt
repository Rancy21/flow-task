package com.rancy21.flowtask.ui.today

import androidx.compose.foundation.clickable
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
import androidx.compose.ui.text.style.TextDecoration
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import com.rancy21.flowtask.data.entity.TaskEntity
import com.rancy21.flowtask.ui.theme.PriorityColors
import org.koin.androidx.compose.koinViewModel

@Composable
fun TodayScreen(
    viewModel: TodayViewModel = koinViewModel(),
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
                "No tasks for today",
                modifier = Modifier.align(Alignment.Center),
                style = MaterialTheme.typography.bodyLarge,
                color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.5f),
            )
        } else {
            LazyColumn(
                contentPadding = PaddingValues(16.dp),
                verticalArrangement = Arrangement.spacedBy(8.dp),
            ) {
                items(uiState.tasks, key = { it.id }) { task ->
                    TaskCard(
                        task = task,
                        onMarkDone = { viewModel.markDone(task) },
                        onUnSchedule = { viewModel.unSchedule(task) },
                        onDelete = { viewModel.delete(task) },
                        onClickEdit = { onOpenTaskEditor(task.id) },
                    )
                }
            }
        }
        }
    }
}

@Composable
private fun TaskCard(
    task: TaskEntity,
    onMarkDone: () -> Unit,
    onUnSchedule: () -> Unit,
    onDelete: () -> Unit,
    onClickEdit: () -> Unit,
) {
    val priorityColor = PriorityColors.colorFor(task.priority)

    Card(
        modifier = Modifier.fillMaxWidth(),
        onClick = onClickEdit,
    ) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(12.dp),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            // Priority color bar
            Box(
                modifier = Modifier
                    .width(4.dp)
                    .height(48.dp)
                    .padding(end = 8.dp)
                    .defaultMinSize(minHeight = 48.dp),
            ) {
                Surface(
                    modifier = Modifier.fillMaxSize(),
                    color = priorityColor,
                    shape = MaterialTheme.shapes.small,
                ) {}
            }

            Spacer(modifier = Modifier.width(12.dp))

            Column(modifier = Modifier.weight(1f)) {
                Text(
                    text = task.title,
                    style = MaterialTheme.typography.titleMedium,
                    maxLines = 1,
                    overflow = TextOverflow.Ellipsis,
                )
                if (!task.description.isNullOrBlank()) {
                    Text(
                        text = task.description,
                        style = MaterialTheme.typography.bodySmall,
                        color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.6f),
                        maxLines = 2,
                        overflow = TextOverflow.Ellipsis,
                    )
                }
                Text(
                    text = task.priority,
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
