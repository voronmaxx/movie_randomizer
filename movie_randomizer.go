// movie_randomizer.go
package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

type Movie struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Genre     string `json:"genre"`
	Director  string `json:"director"`
	Year      int    `json:"year"`
	Watched   bool   `json:"watched"`
	Rating    int    `json:"rating"`
	Notes     string `json:"notes"`
	AddedDate string `json:"added_date"`
}

type MoviesData struct {
	Movies []Movie `json:"movies"`
}

type MovieRandomizer struct {
	movies  []Movie
	nextID  int
}

func NewMovieRandomizer() *MovieRandomizer {
	rand.Seed(time.Now().UnixNano())
	return &MovieRandomizer{
		movies:  []Movie{},
		nextID:  1,
	}
}

func (r *MovieRandomizer) AddMovie(title, genre, director string, year int, watched bool, rating int, notes string) (Movie, error) {
	if rating < 1 || rating > 10 {
		return Movie{}, fmt.Errorf("оценка должна быть от 1 до 10")
	}
	currentYear := time.Now().Year()
	if year < 1900 || year > currentYear {
		return Movie{}, fmt.Errorf("год должен быть от 1900 до %d", currentYear)
	}
	if title == "" || genre == "" {
		return Movie{}, fmt.Errorf("название и жанр не могут быть пустыми")
	}
	if director == "" {
		director = "Неизвестен"
	}
	movie := Movie{
		ID:        r.nextID,
		Title:     title,
		Genre:     genre,
		Director:  director,
		Year:      year,
		Watched:   watched,
		Rating:    rating,
		Notes:     notes,
		AddedDate: time.Now().Format("2006-01-02"),
	}
	r.movies = append(r.movies, movie)
	r.nextID++
	return movie, nil
}

func (r *MovieRandomizer) FindMovie(id int) *Movie {
	for i := range r.movies {
		if r.movies[i].ID == id {
			return &r.movies[i]
		}
	}
	return nil
}

func (r *MovieRandomizer) EditMovie(id int, updates map[string]interface{}) bool {
	movie := r.FindMovie(id)
	if movie == nil {
		return false
	}
	for key, value := range updates {
		switch key {
		case "title":
			if v, ok := value.(string); ok {
				movie.Title = v
			}
		case "genre":
			if v, ok := value.(string); ok {
				movie.Genre = v
			}
		case "director":
			if v, ok := value.(string); ok {
				movie.Director = v
			}
		case "year":
			if v, ok := value.(int); ok {
				movie.Year = v
			}
		case "watched":
			if v, ok := value.(bool); ok {
				movie.Watched = v
			}
		case "rating":
			if v, ok := value.(int); ok {
				movie.Rating = v
			}
		case "notes":
			if v, ok := value.(string); ok {
				movie.Notes = v
			}
		}
	}
	return true
}

func (r *MovieRandomizer) DeleteMovie(id int) bool {
	for i, m := range r.movies {
		if m.ID == id {
			r.movies = append(r.movies[:i], r.movies[i+1:]...)
			return true
		}
	}
	return false
}

func (r *MovieRandomizer) GetUnwatched() []Movie {
	var result []Movie
	for _, m := range r.movies {
		if !m.Watched {
			result = append(result, m)
		}
	}
	return result
}

func (r *MovieRandomizer) GetByGenre(genre string) []Movie {
	var result []Movie
	for _, m := range r.movies {
		if strings.EqualFold(m.Genre, genre) {
			result = append(result, m)
		}
	}
	return result
}

func (r *MovieRandomizer) RandomMovie(genre string) *Movie {
	pool := r.GetUnwatched()
	if genre != "" {
		pool = r.GetByGenre(genre)
	}
	if len(pool) == 0 {
		// Если нет подходящих, берём все (или по жанру)
		if genre != "" {
			pool = r.GetByGenre(genre)
		} else {
			pool = r.movies
		}
	}
	if len(pool) == 0 {
		return nil
	}
	return &pool[rand.Intn(len(pool))]
}

func (r *MovieRandomizer) GetStats() map[string]interface{} {
	total := len(r.movies)
	watched := 0
	for _, m := range r.movies {
		if m.Watched {
			watched++
		}
	}
	unwatched := total - watched
	var ratings []int
	for _, m := range r.movies {
		if m.Watched {
			ratings = append(ratings, m.Rating)
		}
	}
	avgRating := 0.0
	if len(ratings) > 0 {
		sum := 0
		for _, r := range ratings {
			sum += r
		}
		avgRating = float64(sum) / float64(len(ratings))
	}
	genres := make(map[string]int)
	for _, m := range r.movies {
		genres[m.Genre]++
	}
	return map[string]interface{}{
		"total":       total,
		"watched":     watched,
		"unwatched":   unwatched,
		"avg_rating":  avgRating,
		"genres":      genres,
	}
}

func (r *MovieRandomizer) SaveToFile(filename string) error {
	data := MoviesData{Movies: r.movies}
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, jsonData, 0644)
}

func (r *MovieRandomizer) LoadFromFile(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	var md MoviesData
	if err := json.Unmarshal(data, &md); err != nil {
		return err
	}
	r.movies = md.Movies
	for _, m := range r.movies {
		if m.ID >= r.nextID {
			r.nextID = m.ID + 1
		}
	}
	return nil
}

func (r *MovieRandomizer) ExportCSV(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	writer := csv.NewWriter(file)
	writer.Comma = ';'
	defer writer.Flush()
	headers := []string{"ID", "Название", "Жанр", "Режиссёр", "Год", "Просмотрен", "Оценка", "Заметки", "Дата добавления"}
	if err := writer.Write(headers); err != nil {
		return err
	}
	for _, m := range r.movies {
		watchedStr := "Нет"
		if m.Watched {
			watchedStr = "Да"
		}
		row := []string{
			strconv.Itoa(m.ID),
			m.Title,
			m.Genre,
			m.Director,
			strconv.Itoa(m.Year),
			watchedStr,
			strconv.Itoa(m.Rating),
			m.Notes,
			m.AddedDate,
		}
		if err := writer.Write(row); err != nil {
			return err
		}
	}
	return nil
}

func (r *MovieRandomizer) ImportCSV(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	reader := csv.NewReader(file)
	reader.Comma = ';'
	reader.LazyQuotes = true
	records, err := reader.ReadAll()
	if err != nil {
		return err
	}
	if len(records) < 2 {
		return fmt.Errorf("файл пуст или нет данных")
	}
	for _, row := range records[1:] {
		if len(row) < 9 {
			continue
		}
		title := row[1]
		genre := row[2]
		director := row[3]
		year, _ := strconv.Atoi(row[4])
		watched := row[5] == "Да"
		rating, _ := strconv.Atoi(row[6])
		notes := row[7]
		_, err := r.AddMovie(title, genre, director, year, watched, rating, notes)
		if err != nil {
			fmt.Println("Ошибка импорта строки:", err)
		}
	}
	return nil
}

func readString(prompt string) string {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

func readInt(prompt string) int {
	for {
		input := readString(prompt)
		if val, err := strconv.Atoi(input); err == nil {
			return val
		}
		fmt.Println("Введите число.")
	}
}

func readBool(prompt string) bool {
	for {
		input := readString(prompt)
		if input == "1" {
			return true
		} else if input == "0" {
			return false
		}
		fmt.Println("Введите 1 или 0.")
	}
}

func printMovie(movie Movie) {
	status := "✅ Просмотрен"
	if !movie.Watched {
		status = "⏳ Не просмотрен"
	}
	fmt.Printf("#%d - %s (%d)\n", movie.ID, movie.Title, movie.Year)
	fmt.Printf("   Жанр: %s, Режиссёр: %s\n", movie.Genre, movie.Director)
	fmt.Printf("   %s, Оценка: %d/10\n", status, movie.Rating)
	if movie.Notes != "" {
		fmt.Printf("   Заметки: %s\n", movie.Notes)
	}
	fmt.Printf("   Добавлен: %s\n", movie.AddedDate)
}

func main() {
	randomizer := NewMovieRandomizer()
	if err := randomizer.LoadFromFile("movies_data.json"); err != nil {
		fmt.Println("Ошибка загрузки:", err)
	}

	for {
		fmt.Println("\n===== ГЕНЕРАТОР СЛУЧАЙНОГО ФИЛЬМА (Go) =====")
		fmt.Println("1. Добавить фильм")
		fmt.Println("2. Показать все фильмы")
		fmt.Println("3. Рекомендовать случайный фильм")
		fmt.Println("4. Рекомендовать по жанру")
		fmt.Println("5. Показать непросмотренные фильмы")
		fmt.Println("6. Отметить фильм как просмотренный")
		fmt.Println("7. Редактировать фильм")
		fmt.Println("8. Удалить фильм")
		fmt.Println("9. Показать статистику")
		fmt.Println("10. Сохранить в файл")
		fmt.Println("11. Загрузить из файла")
		fmt.Println("12. Экспорт в CSV")
		fmt.Println("13. Импорт из CSV")
		fmt.Println("0. Выход")
		choice := readString("Выберите действие: ")

		switch choice {
		case "0":
			return
		case "1":
			title := readString("Название: ")
			if title == "" {
				fmt.Println("Название не может быть пустым.")
				continue
			}
			genre := readString("Жанр: ")
			if genre == "" {
				fmt.Println("Жанр не может быть пустым.")
				continue
			}
			director := readString("Режиссёр (необязательно): ")
			year := readInt("Год выпуска: ")
			watched := readBool("Статус (1-просмотрен, 0-нет): ")
			rating := readInt("Оценка (1-10): ")
			notes := readString("Заметки (необязательно): ")
			movie, err := randomizer.AddMovie(title, genre, director, year, watched, rating, notes)
			if err != nil {
				fmt.Println("Ошибка:", err)
			} else {
				fmt.Printf("Фильм добавлен с ID %d\n", movie.ID)
			}
		case "2":
			if len(randomizer.movies) == 0 {
				fmt.Println("Нет фильмов.")
			} else {
				for _, m := range randomizer.movies {
					printMovie(m)
				}
			}
		case "3":
			movie := randomizer.RandomMovie("")
			if movie == nil {
				fmt.Println("Нет фильмов в коллекции.")
			} else {
				fmt.Println("\n🎬 Рекомендую посмотреть:")
				printMovie(*movie)
			}
		case "4":
			genre := readString("Введите жанр: ")
			if genre == "" {
				fmt.Println("Введите жанр.")
				continue
			}
			movie := randomizer.RandomMovie(genre)
			if movie == nil {
				fmt.Printf("Нет фильмов в жанре '%s'.\n", genre)
			} else {
				fmt.Printf("\n🎬 Рекомендую посмотреть в жанре %s:\n", genre)
				printMovie(*movie)
			}
		case "5":
			unwatched := randomizer.GetUnwatched()
			if len(unwatched) == 0 {
				fmt.Println("Нет непросмотренных фильмов.")
			} else {
				for _, m := range unwatched {
					printMovie(m)
				}
			}
		case "6":
			id := readInt("Введите ID фильма: ")
			movie := randomizer.FindMovie(id)
			if movie == nil {
				fmt.Println("Фильм не найден.")
				continue
			}
			if movie.Watched {
				fmt.Println("Фильм уже отмечен как просмотренный.")
			} else {
				movie.Watched = true
				fmt.Println("Фильм отмечен как просмотренный.")
			}
		case "7":
			id := readInt("Введите ID фильма для редактирования: ")
			movie := randomizer.FindMovie(id)
			if movie == nil {
				fmt.Println("Фильм не найден.")
				continue
			}
			fmt.Println("Оставьте поле пустым, чтобы не менять.")
			newTitle := readString(fmt.Sprintf("Название (%s): ", movie.Title))
			newGenre := readString(fmt.Sprintf("Жанр (%s): ", movie.Genre))
			newDirector := readString(fmt.Sprintf("Режиссёр (%s): ", movie.Director))
			newYear := readString(fmt.Sprintf("Год (%d): ", movie.Year))
			newWatched := readString(fmt.Sprintf("Статус (1-просмотрен, 0-нет) сейчас: %d: ", map[bool]int{true: 1, false: 0}[movie.Watched]))
			newRating := readString(fmt.Sprintf("Оценка (%d): ", movie.Rating))
			newNotes := readString(fmt.Sprintf("Заметки (%s): ", movie.Notes))
			updates := make(map[string]interface{})
			if newTitle != "" {
				updates["title"] = newTitle
			}
			if newGenre != "" {
				updates["genre"] = newGenre
			}
			if newDirector != "" {
				updates["director"] = newDirector
			}
			if newYear != "" {
				if y, err := strconv.Atoi(newYear); err == nil {
					updates["year"] = y
				} else {
					fmt.Println("Год должен быть числом, пропускаем.")
				}
			}
			if newWatched != "" {
				updates["watched"] = newWatched == "1"
			}
			if newRating != "" {
				if r, err := strconv.Atoi(newRating); err == nil {
					updates["rating"] = r
				} else {
					fmt.Println("Оценка должна быть числом, пропускаем.")
				}
			}
			if newNotes != "" {
				updates["notes"] = newNotes
			}
			if randomizer.EditMovie(id, updates) {
				fmt.Println("Фильм обновлён.")
			} else {
				fmt.Println("Ошибка обновления.")
			}
		case "8":
			id := readInt("Введите ID фильма для удаления: ")
			if randomizer.DeleteMovie(id) {
				fmt.Println("Фильм удалён.")
			} else {
				fmt.Println("Фильм не найден.")
			}
		case "9":
			stats := randomizer.GetStats()
			fmt.Println("\n=== СТАТИСТИКА ===")
			fmt.Printf("Всего фильмов: %d\n", stats["total"])
			fmt.Printf("Просмотрено: %d\n", stats["watched"])
			fmt.Printf("Не просмотрено: %d\n", stats["unwatched"])
			fmt.Printf("Средняя оценка (просмотренные): %.2f\n", stats["avg_rating"])
			fmt.Println("По жанрам:")
			genres := stats["genres"].(map[string]int)
			for g, c := range genres {
				fmt.Printf("  %s: %d\n", g, c)
			}
		case "10":
			if err := randomizer.SaveToFile("movies_data.json"); err != nil {
				fmt.Println("Ошибка сохранения:", err)
			} else {
				fmt.Println("Сохранено.")
			}
		case "11":
			if err := randomizer.LoadFromFile("movies_data.json"); err != nil {
				fmt.Println("Ошибка загрузки:", err)
			} else {
				fmt.Println("Загружено.")
			}
		case "12":
			if err := randomizer.ExportCSV("movies_export.csv"); err != nil {
				fmt.Println("Ошибка экспорта:", err)
			} else {
				fmt.Println("Экспортировано в movies_export.csv")
			}
		case "13":
			if err := randomizer.ImportCSV("movies_export.csv"); err != nil {
				fmt.Println("Ошибка импорта:", err)
			} else {
				fmt.Println("Импортировано из movies_export.csv")
			}
		default:
			fmt.Println("Неизвестная команда.")
		}
	}
}
