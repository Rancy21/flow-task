package com.rancy21.flowtask.ui.editor

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.Add
import androidx.compose.material.icons.filled.Check
import androidx.compose.material.icons.filled.Close
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import com.rancy21.flowtask.data.entity.NoteEntity
import com.rancy21.flowtask.ui.theme.PriorityColors
import kotlinx.coroutines.launch
import org.koin.androidx.compose.koinViewModel
import java.time.LocalDate
import java.time.format.DateTimeFormatter

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun TaskEditorScreen(
    taskId: String? = null,
    onClose: () -> Unit,
    viewModel: TaskEditorViewModel = koinViewModel(),
) {
    LaunchedEffect(taskId) {
        if (taskId != null) {
            viewModel.loadTask(taskId)
        } else {
            viewModel.reset()
        }
    }

    val uiState by viewModel.uiState.collectAsState()
    val snackbarHostState = remember { SnackbarHostState() }
    val scope = rememberCoroutineScope()
    var showDatePicker by remember { mutableStateOf(false) }

    if (showDatePicker) {
        val datePickerState = rememberDatePickerState(
            initialSelectedDateMillis = uiState.scheduledDate?.let {
                LocalDate.parse(it).toEpochDay() * 86400000L
            }
        )
        DatePickerDialog(
            onDismissRequest = { showDatePicker = false },
            confirmButton = {
                TextButton(onClick = {
                    val millis = datePickerState.selectedDateMillis
                    if (millis != null) {
                        val date = java.time.Instant.ofEpochMilli(millis)
                            .atZone(java.time.ZoneId.systemDefault())
                            .toLocalDate()
                        viewModel.setScheduledDate(date.format(DateTimeFormatter.ISO_LOCAL_DATE))
                    }
                    showDatePicker = false
                }) { Text("OK") }
            },
            dismissButton = {
                TextButton(onClick = { showDatePicker = false }) { Text("Cancel") }
            },
        ) {
            DatePicker(state = datePickerState)
        }
    }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text(if (uiState.isNew) "New Task" else "Edit Task") },
                navigationIcon = {
                    IconButton(onClick = onClose) {
                        Icon(Icons.AutoMirrored.Filled.ArrowBack, contentDescription = "Close")
                    }
                },
                actions = {
                    IconButton(
                        onClick = {
                            if (viewModel.save()) {
                                onClose()
                            } else {
                                scope.launch {
                                    snackbarHostState.showSnackbar("Title is required")
                                }
                            }
                        },
                        enabled = !uiState.isSaving,
                    ) {
                        Icon(Icons.Filled.Check, contentDescription = "Save")
                    }
                },
            )
        },
        snackbarHost = { SnackbarHost(snackbarHostState) },
    ) { innerPadding ->
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(innerPadding)
                .imePadding()
                .verticalScroll(rememberScrollState()),
        ) {
            Column(
                modifier = Modifier.padding(16.dp),
                verticalArrangement = Arrangement.spacedBy(16.dp),
            ) {
                // Title
                OutlinedTextField(
                    value = uiState.title,
                    onValueChange = { viewModel.setTitle(it) },
                    label = { Text("Title") },
                    singleLine = true,
                    modifier = Modifier.fillMaxWidth(),
                )

                // Description
                OutlinedTextField(
                    value = uiState.description,
                    onValueChange = { viewModel.setDescription(it) },
                    label = { Text("Description") },
                    minLines = 3,
                    maxLines = 6,
                    modifier = Modifier.fillMaxWidth(),
                )

                // Priority
                Text("Priority", style = MaterialTheme.typography.labelLarge)
                Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                    listOf("P1", "P2", "P3").forEach { p ->
                        FilterChip(
                            selected = uiState.priority == p,
                            onClick = { viewModel.setPriority(p) },
                            label = { Text(p) },
                            colors = FilterChipDefaults.filterChipColors(
                                selectedContainerColor = PriorityColors.colorFor(p).copy(alpha = 0.2f),
                                selectedLabelColor = PriorityColors.colorFor(p),
                            ),
                        )
                    }
                }

                // Date
                Text("Scheduled Date", style = MaterialTheme.typography.labelLarge)
                Row(
                    verticalAlignment = Alignment.CenterVertically,
                    horizontalArrangement = Arrangement.spacedBy(8.dp),
                ) {
                    OutlinedButton(onClick = { showDatePicker = true }) {
                        Text(
                            uiState.scheduledDate?.let { formatDate(it) } ?: "Pick a date"
                        )
                    }
                    if (uiState.scheduledDate != null) {
                        TextButton(onClick = { viewModel.setScheduledDate(null) }) {
                            Text("Clear")
                        }
                    }
                }

                if (uiState.isSaving) {
                    LinearProgressIndicator(modifier = Modifier.fillMaxWidth())
                }

                // Notes section — only for existing tasks
                if (!uiState.isNew) {
                    HorizontalDivider(modifier = Modifier.padding(vertical = 8.dp))
                    Text("Notes", style = MaterialTheme.typography.titleSmall)

                    // Existing notes
                    uiState.notes.forEach { note ->
                        NoteRow(
                            note = note,
                            onDelete = { viewModel.deleteNote(note) },
                        )
                    }

                    // Add note input
                    Row(
                        modifier = Modifier.fillMaxWidth(),
                        verticalAlignment = Alignment.Bottom,
                        horizontalArrangement = Arrangement.spacedBy(8.dp),
                    ) {
                        OutlinedTextField(
                            value = uiState.newNoteContent,
                            onValueChange = { viewModel.setNewNoteContent(it) },
                            label = { Text("Add a reflection…") },
                            modifier = Modifier.weight(1f),
                            maxLines = 3,
                        )
                        IconButton(onClick = { viewModel.addNote() }) {
                            Icon(Icons.Filled.Add, contentDescription = "Add note")
                        }
                    }
                }
            }
        }
    }
}

@Composable
private fun NoteRow(note: NoteEntity, onDelete: () -> Unit) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .padding(vertical = 4.dp),
        verticalAlignment = Alignment.Top,
    ) {
        Text(
            note.content,
            style = MaterialTheme.typography.bodySmall,
            modifier = Modifier.weight(1f),
            maxLines = 3,
            overflow = TextOverflow.Ellipsis,
        )
        IconButton(onClick = onDelete, modifier = Modifier.size(32.dp)) {
            Icon(Icons.Filled.Close, contentDescription = "Delete note", modifier = Modifier.size(16.dp))
        }
    }
}

private fun formatDate(isoDate: String): String {
    return try {
        val date = LocalDate.parse(isoDate)
        date.format(DateTimeFormatter.ofPattern("EEE, MMM d, yyyy"))
    } catch (_: Exception) {
        isoDate
    }
}
