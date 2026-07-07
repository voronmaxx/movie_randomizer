// MovieRandomizer.java
import java.io.*;
import java.nio.file.*;
import java.time.LocalDate;
import java.util.*;
import java.util.stream.Collectors;

record Movie(int id, String title, String genre, String director, int year, boolean watched, int rating, String notes, String addedDate) implements Serializable {}

class MoviesData implements Serializable {
    private static final long serialVersionUID = 1L;
    public List<Movie> movies;
}

class MovieRandomizer implements Serializable {
    private static final long serialVersionUID = 1L;
    private List<Movie> movies = new ArrayList<>();
    private int nextId = 1;

    public Movie addMovie(String title, String genre, String director, int year, boolean watched, int rating, String notes) {
        if (rating < 1 || rating > 10) throw new IllegalArgumentException("Оценка должна быть от 1 до 10");
        if (year < 1900 || year > LocalDate.now().getYear())
            throw new IllegalArgumentException("Год должен быть от 1900 до " + LocalDate.now().getYear());
        if (title.isBlank() || genre.isBlank())
            throw new IllegalArgumentException("Название и жанр не могут быть пустыми");
        if (director.isBlank()) director = "Неизвестен";
        if (notes == null) notes = "";
        Movie movie = new Movie(nextId, title, genre, director, year, watched, rating, notes, LocalDate.now().toString());
        movies.add(movie);
        nextId++;
        return movie;
    }

    public Optional<Movie> findMovie(int id) {
        return movies.stream().filter(m -> m.id() == id).findFirst();
    }

    public boolean editMovie(int id, Map<String, Object> updates) {
        Optional<Movie> opt = findMovie(id);
        if (opt.isEmpty()) return false;
        Movie old = opt.get();
        movies.remove(old);
        String title = (String) updates.getOrDefault("title", old.title());
        String genre = (String) updates.getOrDefault("genre", old.genre());
        String director = (String) updates.getOrDefault("director", old.director());
        int year = (int) updates.getOrDefault("year", old.year());
        boolean watched = (boolean) updates.getOrDefault("watched", old.watched());
        int rating = (int) updates.getOrDefault("rating", old.rating());
        String notes = (String) updates.getOrDefault("notes", old.notes());
        Movie updated = new Movie(old.id(), title, genre, director, year, watched, rating, notes, old.addedDate());
        movies.add(updated);
        return true;
    }

    public boolean deleteMovie(int id) {
        return movies.removeIf(m -> m.id() == id);
    }

    public List<Movie> getUnwatched() {
        return movies.stream().filter(m -> !m.watched()).collect(Collectors.toList());
    }

    public List<Movie> getByGenre(String genre) {
        return movies.stream().filter(m -> m.genre().equalsIgnoreCase(genre)).collect(Collectors.toList());
    }

    public Movie randomMovie(String genre) {
        List<Movie> pool = genre != null ? getByGenre(genre) : getUnwatched();
        if (pool.isEmpty()) {
            // если нет, берём все (или по жанру)
            pool = genre != null ? getByGenre(genre) : movies;
        }
        if (pool.isEmpty()) return null;
        Random rand = new Random();
        return pool.get(rand.nextInt(pool.size()));
    }

    public Map<String, Object> getStats() {
        int total = movies.size();
        int watchedCount = (int) movies.stream().filter(Movie::watched).count();
        int unwatched = total - watchedCount;
        double avgRating = movies.stream().filter(Movie::watched).mapToInt(Movie::rating).average().orElse(0);
        Map<String, Integer> genres = new HashMap<>();
        movies.forEach(m -> genres.put(m.genre(), genres.getOrDefault(m.genre(), 0) + 1));
        Map<String, Object> stats = new HashMap<>();
        stats.put("total", total);
        stats.put("watched", watchedCount);
        stats.put("unwatched", unwatched);
        stats.put("avg_rating", avgRating);
        stats.put("genres", genres);
        return stats;
    }

    public void saveToFile(String filename) throws IOException {
        MoviesData data = new MoviesData();
        data.movies = new ArrayList<>(movies);
        try (ObjectOutputStream oos = new ObjectOutputStream(Files.newOutputStream(Paths.get(filename)))) {
            oos.writeObject(data);
        }
    }

    public void loadFromFile(String filename) throws IOException, ClassNotFoundException {
        try (ObjectInputStream ois = new ObjectInputStream(Files.newInputStream(Paths.get(filename)))) {
            MoviesData data = (MoviesData) ois.readObject();
            movies = new ArrayList<>(data.movies);
            for (Movie m : movies) {
                if (m.id() >= nextId) nextId = m.id() + 1;
            }
        }
    }

    public void exportCSV(String filename) throws IOException {
        try (PrintWriter pw = new PrintWriter(Files.newBufferedWriter(Paths.get(filename)))) {
            pw.println("ID;Название;Жанр;Режиссёр;Год;Просмотрен;Оценка;Заметки;Дата добавления");
            for (Movie m : movies) {
                pw.printf("%d;%s;%s;%s;%d;%s;%d;%s;%s%n",
                        m.id(), m.title(), m.genre(), m.director(), m.year(),
                        m.watched() ? "Да" : "Нет", m.rating(), m.notes(), m.addedDate());
            }
        }
    }

    public void importCSV(String filename) throws IOException {
        try (BufferedReader br = Files.newBufferedReader(Paths.get(filename))) {
            String line = br.readLine(); // header
            while ((line = br.readLine()) != null) {
                String[] parts = line.split(";");
                if (parts.length < 9) continue;
                String title = parts[1];
                String genre = parts[2];
                String director = parts[3];
                int year = Integer.parseInt(parts[4]);
                boolean watched = parts[5].equals("Да");
                int rating = Integer.parseInt(parts[6]);
                String notes = parts[7];
                try {
                    addMovie(title, genre, director, year, watched, rating, notes);
                } catch (Exception e) {
                    System.out.println("Ошибка импорта строки: " + e.getMessage());
                }
            }
        }
    }

    public List<Movie> getMovies() { return Collections.unmodifiableList(movies); }
}

public class MovieRandomizerApp {
    private static final Scanner scanner = new Scanner(System.in);
    private static final Random rand = new Random();

    private static String readString(String prompt) {
        System.out.print(prompt);
        return scanner.nextLine().trim();
    }

    private static int readInt(String prompt) {
        while (true) {
            try {
                System.out.print(prompt);
                return Integer.parseInt(scanner.nextLine().trim());
            } catch (NumberFormatException e) {
                System.out.println("Введите число.");
            }
        }
    }

    private static boolean readBool(String prompt) {
        while (true) {
            String input = readString(prompt);
            if (input.equals("1")) return true;
            if (input.equals("0")) return false;
            System.out.println("Введите 1 или 0.");
        }
    }

    private static void printMovie(Movie movie) {
        String status = movie.watched() ? "✅ Просмотрен" : "⏳ Не просмотрен";
        System.out.printf("#%d - %s (%d)%n", movie.id(), movie.title(), movie.year());
        System.out.printf("   Жанр: %s, Режиссёр: %s%n", movie.genre(), movie.director());
        System.out.printf("   %s, Оценка: %d/10%n", status, movie.rating());
        if (!movie.notes().isBlank()) {
            System.out.printf("   Заметки: %s%n", movie.notes());
        }
        System.out.printf("   Добавлен: %s%n", movie.addedDate());
    }

    public static void main(String[] args) {
        MovieRandomizer randomizer = new MovieRandomizer();
        try {
            randomizer.loadFromFile("movies_data.ser");
        } catch (IOException | ClassNotFoundException e) {
            System.out.println("Не удалось загрузить данные.");
        }

        while (true) {
            System.out.println("\n===== ГЕНЕРАТОР СЛУЧАЙНОГО ФИЛЬМА (Java) =====");
            System.out.println("1. Добавить фильм");
            System.out.println("2. Показать все фильмы");
            System.out.println("3. Рекомендовать случайный фильм");
            System.out.println("4. Рекомендовать по жанру");
            System.out.println("5. Показать непросмотренные фильмы");
            System.out.println("6. Отметить фильм как просмотренный");
            System.out.println("7. Редактировать фильм");
            System.out.println("8. Удалить фильм");
            System.out.println("9. Показать статистику");
            System.out.println("10. Сохранить в файл");
            System.out.println("11. Загрузить из файла");
            System.out.println("12. Экспорт в CSV");
            System.out.println("13. Импорт из CSV");
            System.out.println("0. Выход");
            String choice = readString("Выберите действие: ");

            switch (choice) {
                case "0" -> { return; }
                case "1" -> {
                    String title = readString("Название: ");
                    if (title.isBlank()) { System.out.println("Название не может быть пустым."); continue; }
                    String genre = readString("Жанр: ");
                    if (genre.isBlank()) { System.out.println("Жанр не может быть пустым."); continue; }
                    String director = readString("Режиссёр (необязательно): ");
                    int year = readInt("Год выпуска: ");
                    boolean watched = readBool("Статус (1-просмотрен, 0-нет): ");
                    int rating = readInt("Оценка (1-10): ");
                    String notes = readString("Заметки (необязательно): ");
                    try {
                        Movie movie = randomizer.addMovie(title, genre, director, year, watched, rating, notes);
                        System.out.println("Фильм добавлен с ID " + movie.id());
                    } catch (Exception e) {
                        System.out.println("Ошибка: " + e.getMessage());
                    }
                }
                case "2" -> {
                    if (randomizer.getMovies().isEmpty()) System.out.println("Нет фильмов.");
                    else randomizer.getMovies().forEach(MovieRandomizerApp::printMovie);
                }
                case "3" -> {
                    Movie movie = randomizer.randomMovie(null);
                    if (movie == null) System.out.println("Нет фильмов в коллекции.");
                    else { System.out.println("\n🎬 Рекомендую посмотреть:"); printMovie(movie); }
                }
                case "4" -> {
                    String genre = readString("Введите жанр: ");
                    if (genre.isBlank()) { System.out.println("Введите жанр."); continue; }
                    Movie movie = randomizer.randomMovie(genre);
                    if (movie == null) System.out.println("Нет фильмов в жанре '" + genre + "'.");
                    else { System.out.println("\n🎬 Рекомендую посмотреть в жанре " + genre + ":"); printMovie(movie); }
                }
                case "5" -> {
                    var unwatched = randomizer.getUnwatched();
                    if (unwatched.isEmpty()) System.out.println("Нет непросмотренных фильмов.");
                    else unwatched.forEach(MovieRandomizerApp::printMovie);
                }
                case "6" -> {
                    int id = readInt("Введите ID фильма: ");
                    var opt = randomizer.findMovie(id);
                    if (opt.isEmpty()) { System.out.println("Фильм не найден."); continue; }
                    Movie old = opt.get();
                    if (old.watched()) System.out.println("Фильм уже отмечен как просмотренный.");
                    else {
                        randomizer.editMovie(id, Map.of("watched", true));
                        System.out.println("Фильм отмечен как просмотренный.");
                    }
                }
                case "7" -> {
                    int id = readInt("Введите ID фильма для редактирования: ");
                    var opt = randomizer.findMovie(id);
                    if (opt.isEmpty()) { System.out.println("Фильм не найден."); continue; }
                    Movie old = opt.get();
                    System.out.println("Оставьте поле пустым, чтобы не менять.");
                    String newTitle = readString("Название (" + old.title() + "): ");
                    String newGenre = readString("Жанр (" + old.genre() + "): ");
                    String newDirector = readString("Режиссёр (" + old.director() + "): ");
                    String newYearStr = readString("Год (" + old.year() + "): ");
                    String newWatchedStr = readString("Статус (1-просмотрен, 0-нет) сейчас: " + (old.watched() ? "1" : "0") + ": ");
                    String newRatingStr = readString("Оценка (" + old.rating() + "): ");
                    String newNotes = readString("Заметки (" + old.notes() + "): ");
                    Map<String, Object> updates = new HashMap<>();
                    if (!newTitle.isBlank()) updates.put("title", newTitle);
                    if (!newGenre.isBlank()) updates.put("genre", newGenre);
                    if (!newDirector.isBlank()) updates.put("director", newDirector);
                    if (!newYearStr.isBlank()) {
                        try { updates.put("year", Integer.parseInt(newYearStr)); }
                        catch (NumberFormatException e) { System.out.println("Год должен быть числом, пропускаем."); }
                    }
                    if (!newWatchedStr.isBlank()) updates.put("watched", newWatchedStr.equals("1"));
                    if (!newRatingStr.isBlank()) {
                        try { updates.put("rating", Integer.parseInt(newRatingStr)); }
                        catch (NumberFormatException e) { System.out.println("Оценка должна быть числом, пропускаем."); }
                    }
                    if (!newNotes.isBlank()) updates.put("notes", newNotes);
                    if (randomizer.editMovie(id, updates)) System.out.println("Фильм обновлён.");
                    else System.out.println("Ошибка обновления.");
                }
                case "8" -> {
                    int id = readInt("Введите ID фильма для удаления: ");
                    if (randomizer.deleteMovie(id)) System.out.println("Фильм удалён.");
                    else System.out.println("Фильм не найден.");
                }
                case "9" -> {
                    var stats = randomizer.getStats();
                    System.out.println("\n=== СТАТИСТИКА ===");
                    System.out.println("Всего фильмов: " + stats.get("total"));
                    System.out.println("Просмотрено: " + stats.get("watched"));
                    System.out.println("Не просмотрено: " + stats.get("unwatched"));
                    System.out.printf("Средняя оценка (просмотренные): %.2f%n", stats.get("avg_rating"));
                    System.out.println("По жанрам:");
                    @SuppressWarnings("unchecked")
                    Map<String, Integer> genres = (Map<String, Integer>) stats.get("genres");
                    genres.forEach((g, c) -> System.out.println("  " + g + ": " + c));
                }
                case "10" -> {
                    try {
                        randomizer.saveToFile("movies_data.ser");
                        System.out.println("Сохранено.");
                    } catch (IOException e) {
                        System.out.println("Ошибка сохранения: " + e.getMessage());
                    }
                }
                case "11" -> {
                    try {
                        randomizer.loadFromFile("movies_data.ser");
                        System.out.println("Загружено.");
                    } catch (IOException | ClassNotFoundException e) {
                        System.out.println("Ошибка загрузки: " + e.getMessage());
                    }
                }
                case "12" -> {
                    try {
                        randomizer.exportCSV("movies_export.csv");
                        System.out.println("Экспортировано в movies_export.csv");
                    } catch (IOException e) {
                        System.out.println("Ошибка экспорта: " + e.getMessage());
                    }
                }
                case "13" -> {
                    try {
                        randomizer.importCSV("movies_export.csv");
                        System.out.println("Импортировано из movies_export.csv");
                    } catch (IOException e) {
                        System.out.println("Ошибка импорта: " + e.getMessage());
                    }
                }
                default -> System.out.println("Неизвестная команда.");
            }
        }
    }
}
