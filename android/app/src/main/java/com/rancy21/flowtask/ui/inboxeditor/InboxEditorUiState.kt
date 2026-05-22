package com.rancy21.flowtask.ui.inboxeditor

data class InboxEditorUiState(
    val title: String = "",
    val description: String = "",
    val isNew: Boolean = true,
    val itemId: String? = null,
    val isSaving: Boolean = false,
)
