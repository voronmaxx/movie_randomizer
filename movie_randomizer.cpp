// movie_randomizer.cpp
#include <iostream>
#include <vector>
#include <string>
#include <fstream>
#include <sstream>
#include <algorithm>
#include <iomanip>
#include <ctime>
#include <map>
#include <variant>
#include <regex>
#include <cctype>
#include <random>

using namespace std;

struct Movie {
    int id;
    string title;
    string genre;
    string director;
    int year;
    bool watched;
    int rating;
    string notes;
    string addedDate;

    Movie(int id, const string& title, const string& genre, const string& director,
          int year, bool watched, int rating, const string& notes = "", const string& addedDate = "")
        : id(id), title(title), genre(genre), director(director), year(year),
          watched(watched), rating(rating), notes(notes), addedDate(addedDate) {
        if (addedDate.empty()) {
            time_t now = time(nullptr);
            tm* tm_now = localtime(&now);
            char buf[11];
            strftime(buf, sizeof(buf), "%Y-%m-%d", tm_now);
            this->addedDate = string(buf);
        }
    }
};

class MovieRandomizer {
private:
    vector<Movie> movies;
    int nextId = 1;
    random_device rd;
    mt19937 rng;

public:
    MovieRandomizer() : rng(rd()) {}

    int currentYear() {
        time_t now = time(nullptr);
        tm* tm_now = localtime(&now);
        return tm_now->tm_year + 1900;
    }

    Movie addMovie(const string& title, const string& genre, const string& director,
                   int year, bool watched, int rating, const string& notes = "") {
        if (rating < 1 || rating > 10) throw invalid_argument("Оценка должна быть от 1 до 10");
        int cy = currentYear();
        if (year < 1900 || year > cy) throw invalid_argument("Год должен быть от 1900 до " + to_string(cy));
        if (title.empty() || genre.empty()) throw invalid_argument("Название и жанр не могут быть пустыми");
        string d = director.empty() ? "Неизвестен" : director;
        Movie movie(nextId, title, genre, d, year, watched, rating, notes);
        movies.push_back(movie);
        nextId++;
        return movie;
    }

    Movie* findMovie(int id) {
        auto it = find_if(movies.begin(), movies.end(), [id](const Movie& m) { return m.id == id; });
        return it != movies.end() ? &(*it) : nullptr;
    }

    bool editMovie(int id, const map<string, string>& updates) {
        Movie* movie = findMovie(id);
        if (!movie) return false;
        for (const auto& [key, value] : updates) {
            if (key == "title") movie->title = value;
            else if (key == "genre") movie->genre = value;
            else if (key == "director") movie->director = value;
            else if (key == "year") movie->year = stoi(value);
            else if (key == "watched") movie->watched = (value == "1");
            else if (key == "rating") movie->rating = stoi(value);
            else if (key == "notes") movie->notes = value;
        }
        return true;
    }

    bool deleteMovie(int id) {
        auto it = find_if(movies.begin(), movies.end(), [id](const Movie& m) { return m.id == id; });
        if (it == movies.end()) return false;
        movies.erase(it);
        return true;
    }

    vector<Movie> getUnwatched() const {
        vector<Movie> result;
        for (const auto& m : movies) {
            if (!m.watched) result.push_back(m);
        }
        return result;
    }

    vector<Movie> getByGenre(const string& genre) const {
        vector<Movie> result;
        for (const auto& m : movies) {
            if (m.genre == genre) result.push_back(m);
        }
        return result;
    }

    Movie* randomMovie(const string& genre = "") {
        vector<Movie> pool = genre.empty() ? getUnwatched() : getByGenre(genre);
        if (pool.empty()) {
            pool = genre.empty() ? movies : getByGenre(genre);
        }
        if (pool.empty()) return nullptr;
        uniform_int_distribution<> dist(0, pool.size() - 1);
        int idx = dist(rng);
        // нужно вернуть указатель на элемент в основном списке, а не на копию
        for (auto& m : movies) {
            if (m.id == pool[idx].id) return &m;
        }
        return nullptr;
    }

    map<string, variant<int, double, map<string, int>>> getStats() const {
        int total = movies.size();
        int watchedCount = 0;
        for (const auto& m : movies) if (m.watched) watchedCount++;
        int unwatched = total - watchedCount;
        int sumRating = 0;
        int ratingCount = 0;
        for (const auto& m : movies) {
            if (m.watched) { sumRating += m.rating; ratingCount++; }
        }
        double avgRating = ratingCount > 0 ? static_cast<double>(sumRating) / ratingCount : 0.0;
        map<string, int> genres;
        for (const auto& m : movies) genres[m.genre]++;
        map<string, variant<int, double, map<string, int>>> stats;
        stats["total"] = total;
        stats["watched"] = watchedCount;
        stats["unwatched"] = unwatched;
        stats["avg_rating"] = avgRating;
        stats["genres"] = genres;
        return stats;
    }

    void saveToFile(const string& filename = "movies_data.txt") {
        ofstream out(filename);
        if (!out) return;
        for (const auto& m : movies) {
            out << m.id << '|'
                << m.title << '|'
                << m.genre << '|'
                << m.director << '|'
                << m.year << '|'
                << m.watched << '|'
                << m.rating << '|'
                << m.notes << '|'
                << m.addedDate << '\n';
        }
    }

    void loadFromFile(const string& filename = "movies_data.txt") {
        ifstream in(filename);
        if (!in) return;
        movies.clear();
        string line;
        while (getline(in, line)) {
            stringstream ss(line);
            string idStr, title, genre, director, yearStr, watchedStr, ratingStr, notes, addedDate;
            getline(ss, idStr, '|');
            getline(ss, title, '|');
            getline(ss, genre, '|');
            getline(ss, director, '|');
            getline(ss, yearStr, '|');
            getline(ss, watchedStr, '|');
            getline(ss, ratingStr, '|');
            getline(ss, notes, '|');
            getline(ss, addedDate, '|');
            int id = stoi(idStr);
            int year = stoi(yearStr);
            bool watched = (watchedStr == "1");
            int rating = stoi(ratingStr);
            movies.emplace_back(id, title, genre, director, year, watched, rating, notes, addedDate);
            if (id >= nextId) nextId = id + 1;
        }
    }

    void exportCSV(const string& filename = "movies_export.csv") {
        ofstream out(filename);
        if (!out) return;
        out << "ID;Название;Жанр;Режиссёр;Год;Просмотрен;Оценка;Заметки;Дата добавления\n";
        for (const auto& m : movies) {
            out << m.id << ';'
                << m.title << ';'
                << m.genre << ';'
                << m.director << ';'
                << m.year << ';'
                << (m.watched ? "Да" : "Нет") << ';'
                << m.rating << ';'
                << m.notes << ';'
                << m.addedDate << '\n';
        }
    }

    void importCSV(const string& filename = "movies_export.csv") {
        ifstream in(filename);
        if (!in) return;
        string line;
        getline(in, line); // header
        while (getline(in, line)) {
            stringstream ss(line);
            string idStr, title, genre, director, yearStr, watchedStr, ratingStr, notes, addedDate;
            getline(ss, idStr, ';');
            getline(ss, title, ';');
            getline(ss, genre, ';');
            getline(ss, director, ';');
            getline(ss, yearStr, ';');
            getline(ss, watchedStr, ';');
            getline(ss, ratingStr, ';');
            getline(ss, notes, ';');
            getline(ss, addedDate, ';');
            try {
                addMovie(title, genre, director, stoi(yearStr),
                         watchedStr == "Да", stoi(ratingStr), notes);
            } catch (const exception& e) {
                cout << "Ошибка импорта строки: " << e.what() << "\n";
            }
        }
    }

    const vector<Movie>& getMovies() const { return movies; }
};

string readString(const string& prompt) {
    cout << prompt;
    string input;
    getline(cin, input);
    return input;
}

int readInt(const string& prompt) {
    while (true) {
        cout << prompt;
        string input;
        getline(cin, input);
        try {
            return stoi(input);
        } catch (...) {
            cout << "Введите число.\n";
        }
    }
}

bool readBool(const string& prompt) {
    while (true) {
        string input = readString(prompt);
        if (input == "1") return true;
        if (input == "0") return false;
        cout << "Введите 1 или 0.\n";
    }
}

void printMovie(const Movie& movie) {
    string status = movie.watched ? "✅ Просмотрен" : "⏳ Не просмотрен";
    cout << "#" << movie.id << " - " << movie.title << " (" << movie.year << ")\n";
    cout << "   Жанр: " << movie.genre << ", Режиссёр: " << movie.director << "\n";
    cout << "   " << status << ", Оценка: " << movie.rating << "/10\n";
    if (!movie.notes.empty()) cout << "   Заметки: " << movie.notes << "\n";
    cout << "   Добавлен: " << movie.addedDate << "\n";
}

int main() {
    MovieRandomizer randomizer;
    randomizer.loadFromFile();

    while (true) {
        cout << "\n===== ГЕНЕРАТОР СЛУЧАЙНОГО ФИЛЬМА (C++) =====" << endl;
        cout << "1. Добавить фильм\n";
        cout << "2. Показать все фильмы\n";
        cout << "3. Рекомендовать случайный фильм\n";
        cout << "4. Рекомендовать по жанру\n";
        cout << "5. Показать непросмотренные фильмы\n";
        cout << "6. Отметить фильм как просмотренный\n";
        cout << "7. Редактировать фильм\n";
        cout << "8. Удалить фильм\n";
        cout << "9. Показать статистику\n";
        cout << "10. Сохранить в файл\n";
        cout << "11. Загрузить из файла\n";
        cout << "12. Экспорт в CSV\n";
        cout << "13. Импорт из CSV\n";
        cout << "0. Выход\n";
        string choice = readString("Выберите действие: ");

        if (choice == "0") break;

        if (choice == "1") {
            string title = readString("Название: ");
            if (title.empty()) { cout << "Название не может быть пустым.\n"; continue; }
            string genre = readString("Жанр: ");
            if (genre.empty()) { cout << "Жанр не может быть пустым.\n"; continue; }
            string director = readString("Режиссёр (необязательно): ");
            int year = readInt("Год выпуска: ");
            bool watched = readBool("Статус (1-просмотрен, 0-нет): ");
            int rating = readInt("Оценка (1-10): ");
            string notes = readString("Заметки (необязательно): ");
            try {
                Movie movie = randomizer.addMovie(title, genre, director, year, watched, rating, notes);
                cout << "Фильм добавлен с ID " << movie.id << "\n";
            } catch (const exception& e) {
                cout << "Ошибка: " << e.what() << "\n";
            }
        } else if (choice == "2") {
            if (randomizer.getMovies().empty()) {
                cout << "Нет фильмов.\n";
            } else {
                for (const auto& m : randomizer.getMovies()) printMovie(m);
            }
        } else if (choice == "3") {
            Movie* movie = randomizer.randomMovie();
            if (!movie) {
                cout << "Нет фильмов в коллекции.\n";
            } else {
                cout << "\n🎬 Рекомендую посмотреть:\n";
                printMovie(*movie);
            }
        } else if (choice == "4") {
            string genre = readString("Введите жанр: ");
            if (genre.empty()) { cout << "Введите жанр.\n"; continue; }
            Movie* movie = randomizer.randomMovie(genre);
            if (!movie) {
                cout << "Нет фильмов в жанре '" << genre << "'.\n";
            } else {
                cout << "\n🎬 Рекомендую посмотреть в жанре " << genre << ":\n";
                printMovie(*movie);
            }
        } else if (choice == "5") {
            auto unwatched = randomizer.getUnwatched();
            if (unwatched.empty()) {
                cout << "Нет непросмотренных фильмов.\n";
            } else {
                for (const auto& m : unwatched) printMovie(m);
            }
        } else if (choice == "6") {
            int id = readInt("Введите ID фильма: ");
            Movie* movie = randomizer.findMovie(id);
            if (!movie) { cout << "Фильм не найден.\n"; continue; }
            if (movie->watched) {
                cout << "Фильм уже отмечен как просмотренный.\n";
            } else {
                movie->watched = true;
                cout << "Фильм отмечен как просмотренный.\n";
            }
        } else if (choice == "7") {
            int id = readInt("Введите ID фильма для редактирования: ");
            Movie* movie = randomizer.findMovie(id);
            if (!movie) { cout << "Фильм не найден.\n"; continue; }
            cout << "Оставьте поле пустым, чтобы не менять.\n";
            string newTitle = readString("Название (" + movie->title + "): ");
            string newGenre = readString("Жанр (" + movie->genre + "): ");
            string newDirector = readString("Режиссёр (" + movie->director + "): ");
            string newYear = readString("Год (" + to_string(movie->year) + "): ");
            string newWatched = readString("Статус (1-просмотрен, 0-нет) сейчас: " + string(movie->watched ? "1" : "0") + ": ");
            string newRating = readString("Оценка (" + to_string(movie->rating) + "): ");
            string newNotes = readString("Заметки (" + movie->notes + "): ");
            map<string, string> updates;
            if (!newTitle.empty()) updates["title"] = newTitle;
            if (!newGenre.empty()) updates["genre"] = newGenre;
            if (!newDirector.empty()) updates["director"] = newDirector;
            if (!newYear.empty()) updates["year"] = newYear;
            if (!newWatched.empty()) updates["watched"] = newWatched;
            if (!newRating.empty()) updates["rating"] = newRating;
            if (!newNotes.empty()) updates["notes"] = newNotes;
            if (randomizer.editMovie(id, updates)) {
                cout << "Фильм обновлён.\n";
            } else {
                cout << "Ошибка обновления.\n";
            }
        } else if (choice == "8") {
            int id = readInt("Введите ID фильма для удаления: ");
            if (randomizer.deleteMovie(id)) {
                cout << "Фильм удалён.\n";
            } else {
                cout << "Фильм не найден.\n";
            }
        } else if (choice == "9") {
            auto stats = randomizer.getStats();
            cout << "\n=== СТАТИСТИКА ===\n";
            cout << "Всего фильмов: " << get<int>(stats["total"]) << "\n";
            cout << "Просмотрено: " << get<int>(stats["watched"]) << "\n";
            cout << "Не просмотрено: " << get<int>(stats["unwatched"]) << "\n";
            cout << "Средняя оценка (просмотренные): " << fixed << setprecision(2) << get<double>(stats["avg_rating"]) << "\n";
            cout << "По жанрам:\n";
            auto genres = get<map<string, int>>(stats["genres"]);
            for (const auto& [g, c] : genres) cout << "  " << g << ": " << c << "\n";
        } else if (choice == "10") {
            randomizer.saveToFile();
            cout << "Сохранено.\n";
        } else if (choice == "11") {
            randomizer.loadFromFile();
            cout << "Загружено.\n";
        } else if (choice == "12") {
            randomizer.exportCSV();
            cout << "Экспортировано в movies_export.csv\n";
        } else if (choice == "13") {
            try {
                randomizer.importCSV();
                cout << "Импортировано из movies_export.csv\n";
            } catch (const exception& e) {
                cout << "Ошибка импорта: " << e.what() << "\n";
            }
        } else {
            cout << "Неизвестная команда.\n";
        }
    }
    return 0;
}
