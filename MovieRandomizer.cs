// MovieRandomizer.cs
using System;
using System.Collections.Generic;
using System.IO;
using System.Linq;
using System.Text.Json;
using System.Text.Json.Serialization;

public record Movie(
    int Id,
    string Title,
    string Genre,
    string Director,
    int Year,
    bool Watched,
    int Rating,
    string Notes,
    string AddedDate
);

public class MoviesData
{
    public List<Movie> Movies { get; set; } = new();
}

public class MovieRandomizer
{
    private List<Movie> movies = new();
    private int nextId = 1;
    private static readonly Random rand = new();

    public IReadOnlyList<Movie> Movies => movies.AsReadOnly();

    public Movie AddMovie(string title, string genre, string director, int year, bool watched, int rating, string notes)
    {
        if (rating < 1 || rating > 10) throw new ArgumentException("Оценка должна быть от 1 до 10");
        if (year < 1900 || year > DateTime.Now.Year)
            throw new ArgumentException($"Год должен быть от 1900 до {DateTime.Now.Year}");
        if (string.IsNullOrWhiteSpace(title) || string.IsNullOrWhiteSpace(genre))
            throw new ArgumentException("Название и жанр не могут быть пустыми");
        if (string.IsNullOrWhiteSpace(director)) director = "Неизвестен";
        if (notes == null) notes = "";
        var movie = new Movie(nextId, title, genre, director, year, watched, rating, notes, DateTime.Now.ToString("yyyy-MM-dd"));
        movies.Add(movie);
        nextId++;
        return movie;
    }

    public Movie? FindMovie(int id) => movies.FirstOrDefault(m => m.Id == id);

    public bool EditMovie(int id, Dictionary<string, object> updates)
    {
        var old = FindMovie(id);
        if (old == null) return false;
        movies.Remove(old);
        string title = updates.ContainsKey("title") ? (string)updates["title"] : old.Title;
        string genre = updates.ContainsKey("genre") ? (string)updates["genre"] : old.Genre;
        string director = updates.ContainsKey("director") ? (string)updates["director"] : old.Director;
        int year = updates.ContainsKey("year") ? (int)updates["year"] : old.Year;
        bool watched = updates.ContainsKey("watched") ? (bool)updates["watched"] : old.Watched;
        int rating = updates.ContainsKey("rating") ? (int)updates["rating"] : old.Rating;
        string notes = updates.ContainsKey("notes") ? (string)updates["notes"] : old.Notes;
        var updated = new Movie(old.Id, title, genre, director, year, watched, rating, notes, old.AddedDate);
        movies.Add(updated);
        return true;
    }

    public bool DeleteMovie(int id) => movies.RemoveAll(m => m.Id == id) > 0;

    public List<Movie> GetUnwatched() => movies.Where(m => !m.Watched).ToList();

    public List<Movie> GetByGenre(string genre) =>
        movies.Where(m => string.Equals(m.Genre, genre, StringComparison.OrdinalIgnoreCase)).ToList();

    public Movie? RandomMovie(string genre = null)
    {
        var pool = string.IsNullOrEmpty(genre) ? GetUnwatched() : GetByGenre(genre);
        if (pool.Count == 0)
        {
            // если нет, берём все (или по жанру)
            pool = string.IsNullOrEmpty(genre) ? movies : GetByGenre(genre);
        }
        if (pool.Count == 0) return null;
        return pool[rand.Next(pool.Count)];
    }

    public Dictionary<string, object> GetStats()
    {
        int total = movies.Count;
        int watchedCount = movies.Count(m => m.Watched);
        int unwatched = total - watchedCount;
        double avgRating = movies.Where(m => m.Watched).Any() ? movies.Where(m => m.Watched).Average(m => m.Rating) : 0;
        var genres = movies.GroupBy(m => m.Genre).ToDictionary(g => g.Key, g => g.Count());
        return new Dictionary<string, object>
        {
            ["total"] = total,
            ["watched"] = watchedCount,
            ["unwatched"] = unwatched,
            ["avg_rating"] = avgRating,
            ["genres"] = genres
        };
    }

    public void SaveToFile(string filename)
    {
        var data = new MoviesData { Movies = movies };
        var options = new JsonSerializerOptions { WriteIndented = true };
        string json = JsonSerializer.Serialize(data, options);
        File.WriteAllText(filename, json);
    }

    public void LoadFromFile(string filename)
    {
        if (!File.Exists(filename)) return;
        string json = File.ReadAllText(filename);
        var data = JsonSerializer.Deserialize<MoviesData>(json);
        if (data != null)
        {
            movies = data.Movies;
            nextId = movies.Any() ? movies.Max(m => m.Id) + 1 : 1;
        }
    }

    public void ExportCSV(string filename)
    {
        using var writer = new StreamWriter(filename);
        writer.WriteLine("ID;Название;Жанр;Режиссёр;Год;Просмотрен;Оценка;Заметки;Дата добавления");
        foreach (var m in movies)
        {
            writer.WriteLine($"{m.Id};{m.Title};{m.Genre};{m.Director};{m.Year};{(m.Watched ? "Да" : "Нет")};{m.Rating};{m.Notes};{m.AddedDate}");
        }
    }

    public void ImportCSV(string filename)
    {
        if (!File.Exists(filename)) throw new FileNotFoundException("Файл не найден");
        using var reader = new StreamReader(filename);
        string header = reader.ReadLine(); // skip header
        while (!reader.EndOfStream)
        {
            string line = reader.ReadLine();
            var parts = line.Split(';');
            if (parts.Length < 9) continue;
            string title = parts[1];
            string genre = parts[2];
            string director = parts[3];
            int year = int.Parse(parts[4]);
            bool watched = parts[5] == "Да";
            int rating = int.Parse(parts[6]);
            string notes = parts[7];
            try
            {
                AddMovie(title, genre, director, year, watched, rating, notes);
            }
            catch (Exception ex)
            {
                Console.WriteLine($"Ошибка импорта строки: {ex.Message}");
            }
        }
    }
}

public static class Program
{
    private static string ReadString(string prompt)
    {
        Console.Write(prompt);
        return Console.ReadLine()?.Trim() ?? "";
    }

    private static int ReadInt(string prompt)
    {
        while (true)
        {
            Console.Write(prompt);
            if (int.TryParse(Console.ReadLine(), out int result))
                return result;
            Console.WriteLine("Введите число.");
        }
    }

    private static bool ReadBool(string prompt)
    {
        while (true)
        {
            string input = ReadString(prompt);
            if (input == "1") return true;
            if (input == "0") return false;
            Console.WriteLine("Введите 1 или 0.");
        }
    }

    private static void PrintMovie(Movie movie)
    {
        string status = movie.Watched ? "✅ Просмотрен" : "⏳ Не просмотрен";
        Console.WriteLine($"#{movie.Id} - {movie.Title} ({movie.Year})");
        Console.WriteLine($"   Жанр: {movie.Genre}, Режиссёр: {movie.Director}");
        Console.WriteLine($"   {status}, Оценка: {movie.Rating}/10");
        if (!string.IsNullOrWhiteSpace(movie.Notes))
            Console.WriteLine($"   Заметки: {movie.Notes}");
        Console.WriteLine($"   Добавлен: {movie.AddedDate}");
    }

    public static void Main()
    {
        var randomizer = new MovieRandomizer();
        try { randomizer.LoadFromFile("movies_data.json"); }
        catch { Console.WriteLine("Не удалось загрузить данные."); }

        while (true)
        {
            Console.WriteLine("\n===== ГЕНЕРАТОР СЛУЧАЙНОГО ФИЛЬМА (C#) =====");
            Console.WriteLine("1. Добавить фильм");
            Console.WriteLine("2. Показать все фильмы");
            Console.WriteLine("3. Рекомендовать случайный фильм");
            Console.WriteLine("4. Рекомендовать по жанру");
            Console.WriteLine("5. Показать непросмотренные фильмы");
            Console.WriteLine("6. Отметить фильм как просмотренный");
            Console.WriteLine("7. Редактировать фильм");
            Console.WriteLine("8. Удалить фильм");
            Console.WriteLine("9. Показать статистику");
            Console.WriteLine("10. Сохранить в файл");
            Console.WriteLine("11. Загрузить из файла");
            Console.WriteLine("12. Экспорт в CSV");
            Console.WriteLine("13. Импорт из CSV");
            Console.WriteLine("0. Выход");
            string choice = ReadString("Выберите действие: ");

            switch (choice)
            {
                case "0": return;
                case "1":
                    string title = ReadString("Название: ");
                    if (string.IsNullOrWhiteSpace(title)) { Console.WriteLine("Название не может быть пустым."); continue; }
                    string genre = ReadString("Жанр: ");
                    if (string.IsNullOrWhiteSpace(genre)) { Console.WriteLine("Жанр не может быть пустым."); continue; }
                    string director = ReadString("Режиссёр (необязательно): ");
                    int year = ReadInt("Год выпуска: ");
                    bool watched = ReadBool("Статус (1-просмотрен, 0-нет): ");
                    int rating = ReadInt("Оценка (1-10): ");
                    string notes = ReadString("Заметки (необязательно): ");
                    try
                    {
                        var movie = randomizer.AddMovie(title, genre, director, year, watched, rating, notes);
                        Console.WriteLine($"Фильм добавлен с ID {movie.Id}");
                    }
                    catch (Exception ex) { Console.WriteLine($"Ошибка: {ex.Message}"); }
                    break;
                case "2":
                    if (!randomizer.Movies.Any()) Console.WriteLine("Нет фильмов.");
                    else foreach (var m in randomizer.Movies) PrintMovie(m);
                    break;
                case "3":
                    var movie = randomizer.RandomMovie();
                    if (movie == null) Console.WriteLine("Нет фильмов в коллекции.");
                    else { Console.WriteLine("\n🎬 Рекомендую посмотреть:"); PrintMovie(movie); }
                    break;
                case "4":
                    string g = ReadString("Введите жанр: ");
                    if (string.IsNullOrWhiteSpace(g)) { Console.WriteLine("Введите жанр."); continue; }
                    var byGenre = randomizer.RandomMovie(g);
                    if (byGenre == null) Console.WriteLine($"Нет фильмов в жанре '{g}'.");
                    else { Console.WriteLine($"\n🎬 Рекомендую посмотреть в жанре {g}:"); PrintMovie(byGenre); }
                    break;
                case "5":
                    var unwatched = randomizer.GetUnwatched();
                    if (!unwatched.Any()) Console.WriteLine("Нет непросмотренных фильмов.");
                    else foreach (var m in unwatched) PrintMovie(m);
                    break;
                case "6":
                    int id = ReadInt("Введите ID фильма: ");
                    var found = randomizer.FindMovie(id);
                    if (found == null) { Console.WriteLine("Фильм не найден."); continue; }
                    if (found.Watched) Console.WriteLine("Фильм уже отмечен как просмотренный.");
                    else { randomizer.EditMovie(id, new Dictionary<string, object> { ["watched"] = true }); Console.WriteLine("Фильм отмечен как просмотренный."); }
                    break;
                case "7":
                    int eid = ReadInt("Введите ID фильма для редактирования: ");
                    var old = randomizer.FindMovie(eid);
                    if (old == null) { Console.WriteLine("Фильм не найден."); continue; }
                    Console.WriteLine("Оставьте поле пустым, чтобы не менять.");
                    string newTitle = ReadString($"Название ({old.Title}): ");
                    string newGenre = ReadString($"Жанр ({old.Genre}): ");
                    string newDirector = ReadString($"Режиссёр ({old.Director}): ");
                    string newYearStr = ReadString($"Год ({old.Year}): ");
                    string newWatchedStr = ReadString($"Статус (1-просмотрен, 0-нет) сейчас: {(old.Watched ? "1" : "0")}: ");
                    string newRatingStr = ReadString($"Оценка ({old.Rating}): ");
                    string newNotes = ReadString($"Заметки ({old.Notes}): ");
                    var updates = new Dictionary<string, object>();
                    if (!string.IsNullOrWhiteSpace(newTitle)) updates["title"] = newTitle;
                    if (!string.IsNullOrWhiteSpace(newGenre)) updates["genre"] = newGenre;
                    if (!string.IsNullOrWhiteSpace(newDirector)) updates["director"] = newDirector;
                    if (!string.IsNullOrWhiteSpace(newYearStr))
                    {
                        if (int.TryParse(newYearStr, out int y)) updates["year"] = y;
                        else Console.WriteLine("Год должен быть числом, пропускаем.");
                    }
                    if (!string.IsNullOrWhiteSpace(newWatchedStr)) updates["watched"] = newWatchedStr == "1";
                    if (!string.IsNullOrWhiteSpace(newRatingStr))
                    {
                        if (int.TryParse(newRatingStr, out int r)) updates["rating"] = r;
                        else Console.WriteLine("Оценка должна быть числом, пропускаем.");
                    }
                    if (!string.IsNullOrWhiteSpace(newNotes)) updates["notes"] = newNotes;
                    if (randomizer.EditMovie(eid, updates)) Console.WriteLine("Фильм обновлён.");
                    else Console.WriteLine("Ошибка обновления.");
                    break;
                case "8":
                    int delId = ReadInt("Введите ID фильма для удаления: ");
                    if (randomizer.DeleteMovie(delId)) Console.WriteLine("Фильм удалён.");
                    else Console.WriteLine("Фильм не найден.");
                    break;
                case "9":
                    var stats = randomizer.GetStats();
                    Console.WriteLine("\n=== СТАТИСТИКА ===");
                    Console.WriteLine($"Всего фильмов: {stats["total"]}");
                    Console.WriteLine($"Просмотрено: {stats["watched"]}");
                    Console.WriteLine($"Не просмотрено: {stats["unwatched"]}");
                    Console.WriteLine($"Средняя оценка (просмотренные): {stats["avg_rating"]:F2}");
                    Console.WriteLine("По жанрам:");
                    var genres = (Dictionary<string, int>)stats["genres"];
                    foreach (var kv in genres) Console.WriteLine($"  {kv.Key}: {kv.Value}");
                    break;
                case "10":
                    try { randomizer.SaveToFile("movies_data.json"); Console.WriteLine("Сохранено."); }
                    catch (Exception ex) { Console.WriteLine($"Ошибка: {ex.Message}"); }
                    break;
                case "11":
                    try { randomizer.LoadFromFile("movies_data.json"); Console.WriteLine("Загружено."); }
                    catch (Exception ex) { Console.WriteLine($"Ошибка: {ex.Message}"); }
                    break;
                case "12":
                    try { randomizer.ExportCSV("movies_export.csv"); Console.WriteLine("Экспортировано в movies_export.csv"); }
                    catch (Exception ex) { Console.WriteLine($"Ошибка: {ex.Message}"); }
                    break;
                case "13":
                    try { randomizer.ImportCSV("movies_export.csv"); Console.WriteLine("Импортировано из movies_export.csv"); }
                    catch (Exception ex) { Console.WriteLine($"Ошибка: {ex.Message}"); }
                    break;
                default: Console.WriteLine("Неизвестная команда."); break;
            }
        }
    }
}
