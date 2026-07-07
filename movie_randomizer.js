// movie_randomizer.js
const fs = require('fs').promises;
const readline = require('readline');
const { parse } = require('csv-parse');
const { createObjectCsvWriter } = require('csv-writer');

const rl = readline.createInterface({
    input: process.stdin,
    output: process.stdout
});

const question = (prompt) => new Promise(resolve => rl.question(prompt, resolve));

class Movie {
    constructor(id, title, genre, director, year, watched, rating, notes, addedDate) {
        this.id = id;
        this.title = title;
        this.genre = genre;
        this.director = director || 'Неизвестен';
        this.year = year;
        this.watched = watched;
        this.rating = rating;
        this.notes = notes || '';
        this.addedDate = addedDate || new Date().toISOString().slice(0, 10);
    }
}

class MovieRandomizer {
    constructor() {
        this.movies = [];
        this.nextId = 1;
    }

    addMovie(title, genre, director, year, watched, rating, notes) {
        if (rating < 1 || rating > 10) throw new Error('Оценка должна быть от 1 до 10');
        const currentYear = new Date().getFullYear();
        if (year < 1900 || year > currentYear) throw new Error(`Год должен быть от 1900 до ${currentYear}`);
        if (!title.trim() || !genre.trim()) throw new Error('Название и жанр не могут быть пустыми');
        if (!director.trim()) director = 'Неизвестен';
        const movie = new Movie(this.nextId, title, genre, director, year, watched, rating, notes);
        this.movies.push(movie);
        this.nextId++;
        return movie;
    }

    findMovie(id) {
        return this.movies.find(m => m.id === id);
    }

    editMovie(id, updates) {
        const movie = this.findMovie(id);
        if (!movie) return false;
        Object.assign(movie, updates);
        return true;
    }

    deleteMovie(id) {
        const index = this.movies.findIndex(m => m.id === id);
        if (index === -1) return false;
        this.movies.splice(index, 1);
        return true;
    }

    getUnwatched() {
        return this.movies.filter(m => !m.watched);
    }

    getByGenre(genre) {
        return this.movies.filter(m => m.genre.toLowerCase() === genre.toLowerCase());
    }

    randomMovie(genre) {
        let pool = genre ? this.getByGenre(genre) : this.getUnwatched();
        if (pool.length === 0) {
            // если нет, берём все (или по жанру)
            pool = genre ? this.getByGenre(genre) : this.movies;
        }
        if (pool.length === 0) return null;
        return pool[Math.floor(Math.random() * pool.length)];
    }

    getStats() {
        const total = this.movies.length;
        const watchedCount = this.movies.filter(m => m.watched).length;
        const unwatched = total - watchedCount;
        const ratings = this.movies.filter(m => m.watched).map(m => m.rating);
        const avgRating = ratings.length ? ratings.reduce((a, b) => a + b, 0) / ratings.length : 0;
        const genres = {};
        this.movies.forEach(m => {
            genres[m.genre] = (genres[m.genre] || 0) + 1;
        });
        return { total, watched: watchedCount, unwatched, avgRating, genres };
    }

    async saveToFile(filename = 'movies_data.json') {
        const data = { movies: this.movies };
        await fs.writeFile(filename, JSON.stringify(data, null, 2));
    }

    async loadFromFile(filename = 'movies_data.json') {
        try {
            const data = await fs.readFile(filename, 'utf8');
            const parsed = JSON.parse(data);
            this.movies = parsed.movies.map(m => Object.assign(new Movie(0), m));
            this.nextId = this.movies.reduce((max, m) => Math.max(max, m.id), 0) + 1;
        } catch (err) {
            if (err.code !== 'ENOENT') throw err;
        }
    }

    async exportCSV(filename = 'movies_export.csv') {
        const records = this.movies.map(m => ({
            ID: m.id,
            Название: m.title,
            Жанр: m.genre,
            Режиссёр: m.director,
            Год: m.year,
            Просмотрен: m.watched ? 'Да' : 'Нет',
            Оценка: m.rating,
            Заметки: m.notes,
            'Дата добавления': m.addedDate
        }));
        const csvWriter = createObjectCsvWriter({
            path: filename,
            header: [
                { id: 'ID', title: 'ID' },
                { id: 'Название', title: 'Название' },
                { id: 'Жанр', title: 'Жанр' },
                { id: 'Режиссёр', title: 'Режиссёр' },
                { id: 'Год', title: 'Год' },
                { id: 'Просмотрен', title: 'Просмотрен' },
                { id: 'Оценка', title: 'Оценка' },
                { id: 'Заметки', title: 'Заметки' },
                { id: 'Дата добавления', title: 'Дата добавления' }
            ],
            fieldDelimiter: ';'
        });
        await csvWriter.writeRecords(records);
    }

    async importCSV(filename = 'movies_export.csv') {
        const fileContent = await fs.readFile(filename, 'utf8');
        return new Promise((resolve, reject) => {
            parse(fileContent, { columns: true, delimiter: ';' }, (err, records) => {
                if (err) reject(err);
                for (const row of records) {
                    try {
                        this.addMovie(
                            row['Название'],
                            row['Жанр'],
                            row['Режиссёр'],
                            parseInt(row['Год']),
                            row['Просмотрен'] === 'Да',
                            parseInt(row['Оценка']),
                            row['Заметки']
                        );
                    } catch (e) {
                        console.log('Ошибка импорта строки:', e.message);
                    }
                }
                resolve();
            });
        });
    }
}

function printMovie(movie) {
    const status = movie.watched ? '✅ Просмотрен' : '⏳ Не просмотрен';
    console.log(`#${movie.id} - ${movie.title} (${movie.year})`);
    console.log(`   Жанр: ${movie.genre}, Режиссёр: ${movie.director}`);
    console.log(`   ${status}, Оценка: ${movie.rating}/10`);
    if (movie.notes) console.log(`   Заметки: ${movie.notes}`);
    console.log(`   Добавлен: ${movie.addedDate}`);
}

async function main() {
    const randomizer = new MovieRandomizer();
    await randomizer.loadFromFile();

    while (true) {
        console.log('\n===== ГЕНЕРАТОР СЛУЧАЙНОГО ФИЛЬМА (JavaScript) =====');
        console.log('1. Добавить фильм');
        console.log('2. Показать все фильмы');
        console.log('3. Рекомендовать случайный фильм');
        console.log('4. Рекомендовать по жанру');
        console.log('5. Показать непросмотренные фильмы');
        console.log('6. Отметить фильм как просмотренный');
        console.log('7. Редактировать фильм');
        console.log('8. Удалить фильм');
        console.log('9. Показать статистику');
        console.log('10. Сохранить в файл');
        console.log('11. Загрузить из файла');
        console.log('12. Экспорт в CSV');
        console.log('13. Импорт из CSV');
        console.log('0. Выход');
        const choice = await question('Выберите действие: ');

        if (choice === '0') break;

        switch (choice) {
            case '1': {
                const title = await question('Название: ');
                if (!title.trim()) { console.log('Название не может быть пустым.'); continue; }
                const genre = await question('Жанр: ');
                if (!genre.trim()) { console.log('Жанр не может быть пустым.'); continue; }
                const director = await question('Режиссёр (необязательно): ');
                const year = parseInt(await question('Год выпуска: '));
                const watched = (await question('Статус (1-просмотрен, 0-нет): ')) === '1';
                const rating = parseInt(await question('Оценка (1-10): '));
                const notes = await question('Заметки (необязательно): ');
                try {
                    const movie = randomizer.addMovie(title, genre, director, year, watched, rating, notes);
                    console.log(`Фильм добавлен с ID ${movie.id}`);
                } catch (err) {
                    console.log('Ошибка:', err.message);
                }
                break;
            }
            case '2':
                if (randomizer.movies.length === 0) console.log('Нет фильмов.');
                else randomizer.movies.forEach(printMovie);
                break;
            case '3': {
                const movie = randomizer.randomMovie();
                if (!movie) console.log('Нет фильмов в коллекции.');
                else { console.log('\n🎬 Рекомендую посмотреть:'); printMovie(movie); }
                break;
            }
            case '4': {
                const genre = await question('Введите жанр: ');
                if (!genre.trim()) { console.log('Введите жанр.'); continue; }
                const movie = randomizer.randomMovie(genre);
                if (!movie) console.log(`Нет фильмов в жанре '${genre}'.`);
                else { console.log(`\n🎬 Рекомендую посмотреть в жанре ${genre}:`); printMovie(movie); }
                break;
            }
            case '5': {
                const unwatched = randomizer.getUnwatched();
                if (unwatched.length === 0) console.log('Нет непросмотренных фильмов.');
                else unwatched.forEach(printMovie);
                break;
            }
            case '6': {
                const id = parseInt(await question('Введите ID фильма: '));
                const movie = randomizer.findMovie(id);
                if (!movie) { console.log('Фильм не найден.'); continue; }
                if (movie.watched) console.log('Фильм уже отмечен как просмотренный.');
                else { randomizer.editMovie(id, { watched: true }); console.log('Фильм отмечен как просмотренный.'); }
                break;
            }
            case '7': {
                const id = parseInt(await question('Введите ID фильма для редактирования: '));
                const movie = randomizer.findMovie(id);
                if (!movie) { console.log('Фильм не найден.'); continue; }
                console.log('Оставьте поле пустым, чтобы не менять.');
                const newTitle = await question(`Название (${movie.title}): `);
                const newGenre = await question(`Жанр (${movie.genre}): `);
                const newDirector = await question(`Режиссёр (${movie.director}): `);
                const newYear = await question(`Год (${movie.year}): `);
                const newWatched = await question(`Статус (1-просмотрен, 0-нет) сейчас: ${movie.watched ? '1' : '0'}: `);
                const newRating = await question(`Оценка (${movie.rating}): `);
                const newNotes = await question(`Заметки (${movie.notes}): `);
                const updates = {};
                if (newTitle.trim()) updates.title = newTitle;
                if (newGenre.trim()) updates.genre = newGenre;
                if (newDirector.trim()) updates.director = newDirector;
                if (newYear.trim()) {
                    const y = parseInt(newYear);
                    if (!isNaN(y)) updates.year = y;
                    else console.log('Год должен быть числом, пропускаем.');
                }
                if (newWatched.trim()) updates.watched = newWatched === '1';
                if (newRating.trim()) {
                    const r = parseInt(newRating);
                    if (!isNaN(r)) updates.rating = r;
                    else console.log('Оценка должна быть числом, пропускаем.');
                }
                if (newNotes.trim()) updates.notes = newNotes;
                if (randomizer.editMovie(id, updates)) console.log('Фильм обновлён.');
                else console.log('Ошибка обновления.');
                break;
            }
            case '8': {
                const id = parseInt(await question('Введите ID фильма для удаления: '));
                if (randomizer.deleteMovie(id)) console.log('Фильм удалён.');
                else console.log('Фильм не найден.');
                break;
            }
            case '9': {
                const stats = randomizer.getStats();
                console.log('\n=== СТАТИСТИКА ===');
                console.log(`Всего фильмов: ${stats.total}`);
                console.log(`Просмотрено: ${stats.watched}`);
                console.log(`Не просмотрено: ${stats.unwatched}`);
                console.log(`Средняя оценка (просмотренные): ${stats.avgRating.toFixed(2)}`);
                console.log('По жанрам:');
                for (const [g, c] of Object.entries(stats.genres)) {
                    console.log(`  ${g}: ${c}`);
                }
                break;
            }
            case '10':
                try {
                    await randomizer.saveToFile();
                    console.log('Сохранено.');
                } catch (err) {
                    console.log('Ошибка сохранения:', err.message);
                }
                break;
            case '11':
                try {
                    await randomizer.loadFromFile();
                    console.log('Загружено.');
                } catch (err) {
                    console.log('Ошибка загрузки:', err.message);
                }
                break;
            case '12':
                try {
                    await randomizer.exportCSV();
                    console.log('Экспортировано в movies_export.csv');
                } catch (err) {
                    console.log('Ошибка экспорта:', err.message);
                }
                break;
            case '13':
                try {
                    await randomizer.importCSV();
                    console.log('Импортировано из movies_export.csv');
                } catch (err) {
                    console.log('Ошибка импорта:', err.message);
                }
                break;
            default:
                console.log('Неизвестная команда.');
        }
    }
    rl.close();
}

main().catch(console.error);
