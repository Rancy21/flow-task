package com.rancy21.flowtask.data.repository

import com.rancy21.flowtask.data.dao.NoteDao
import com.rancy21.flowtask.data.entity.NoteEntity
import kotlinx.coroutines.flow.Flow

class NoteRepository(private val noteDao: NoteDao) {
    fun getNotesForTask(taskId: String): Flow<List<NoteEntity>> = noteDao.getNotesForTask(taskId)

    fun getAllNotes(): Flow<List<NoteEntity>> = noteDao.getAllNotes()

    suspend fun getById(id: String): NoteEntity? = noteDao.getById(id)

    suspend fun save(note: NoteEntity) = noteDao.insert(note)

    suspend fun update(note: NoteEntity) = noteDao.update(note)

    suspend fun delete(note: NoteEntity) = noteDao.delete(note)

    suspend fun deleteById(id: String) = noteDao.deleteById(id)

    suspend fun upsert(note: NoteEntity) {
        if (noteDao.getById(note.id) != null) {
            noteDao.update(note)
        } else {
            noteDao.insert(note)
        }
    }
}
