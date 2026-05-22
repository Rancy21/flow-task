package com.rancy21.flowtask.ui.week

import com.rancy21.flowtask.data.entity.TaskEntity

data class WeekUiState(
    val tasks: List<TaskEntity> = emptyList(),
    val isLoading: Boolean = true,
)
