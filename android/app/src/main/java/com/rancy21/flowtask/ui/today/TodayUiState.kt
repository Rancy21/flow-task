package com.rancy21.flowtask.ui.today

import com.rancy21.flowtask.data.entity.TaskEntity

data class TodayUiState(
    val tasks: List<TaskEntity> = emptyList(),
    val isLoading: Boolean = true,
)
