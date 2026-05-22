package com.rancy21.flowtask.ui.editor

import com.rancy21.flowtask.data.entity.NoteEntity

data class TaskEditorUiState(
    val title: String = "",
    val description: String = "",
    val priority: String = "P2",
    val scheduledDate: String? = null,
    val isNew: Boolean = true,
    val taskId: String? = null,
    val isSaving: Boolean = false,
    val notes: List<NoteEntity> = emptyList(),
    val newNoteContent: String = "",
)
