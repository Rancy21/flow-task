package com.rancy21.flowtask.ui.inbox

import com.rancy21.flowtask.data.entity.InboxEntity

data class InboxUiState(
    val items: List<InboxEntity> = emptyList(),
    val isLoading: Boolean = true,
)
