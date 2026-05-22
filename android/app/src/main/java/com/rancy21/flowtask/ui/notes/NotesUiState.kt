package com.rancy21.flowtask.ui.notes

import com.rancy21.flowtask.data.entity.NoteEntity

data class NoteWithTask(
    val note: NoteEntity,
    val taskTitle: String = "Unknown task",
)

data class NotesUiState(
    val notes: List<NoteWithTask> = emptyList(),
    val isLoading: Boolean = true,
)
