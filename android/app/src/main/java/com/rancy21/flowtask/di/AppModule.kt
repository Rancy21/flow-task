package com.rancy21.flowtask.di

import com.rancy21.flowtask.data.database.FlowTaskDatabase
import com.rancy21.flowtask.data.repository.InboxRepository
import com.rancy21.flowtask.data.repository.NoteRepository
import com.rancy21.flowtask.data.repository.TaskRepository
import com.rancy21.flowtask.ui.editor.TaskEditorViewModel
import com.rancy21.flowtask.ui.inbox.InboxViewModel
import com.rancy21.flowtask.ui.inboxeditor.InboxEditorViewModel
import com.rancy21.flowtask.ui.notes.NotesViewModel
import com.rancy21.flowtask.ui.today.TodayViewModel
import com.rancy21.flowtask.ui.week.WeekViewModel
import org.koin.android.ext.koin.androidContext
import org.koin.core.module.dsl.viewModel
import org.koin.dsl.module

val appModule = module {
    single { FlowTaskDatabase.getDatabase(androidContext()) }
    single { get<FlowTaskDatabase>().taskDao() }
    single { get<FlowTaskDatabase>().noteDao() }
    single { get<FlowTaskDatabase>().inboxDao() }

    single { TaskRepository(get()) }
    single { NoteRepository(get()) }
    single { InboxRepository(get()) }

    viewModel { TodayViewModel(get()) }
    viewModel { WeekViewModel(get()) }
    viewModel { InboxViewModel(get()) }
    viewModel { NotesViewModel(get(), get()) }
    viewModel { TaskEditorViewModel(get(), get()) }
    viewModel { InboxEditorViewModel(get()) }
}
