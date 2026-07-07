# movie_randomizer.py
import json
import csv
import random
from dataclasses import dataclass, asdict
from datetime import date
from typing import List, Optional
from pathlib import Path

@dataclass
class Movie:
    id: int
    title: str
    genre: str
    director: str
    year: int
    watched: bool
    rating: int  # 1-10
    notes: str
    added_date: str

class MovieRandomizer:
    def __init__(self):
        self.movies: List[Movie] = []
        self.next_id = 1

    def add_movie(self, title: str, genre: str, director: str, year: int,
                  watched: bool, rating: int, notes: str = "") -> Movie:
        if rating < 1 or rating > 10:
            raise ValueError("Оценка должна быть от 1 до 10")
        if year < 1900 or year > date.today().year:
            raise ValueError(f"Год должен быть от 1900 до {date.today().year}")
        if not title or not genre:
            raise ValueError("Название и жанр не могут быть пустыми")
        movie = Movie(
            id=self.next_id,
            title=title,
            genre=genre,
            director=director or "Неизвестен",
            year=year,
            watched=watched,
            rating=rating,
            notes=notes,
            added_date=date.today().isoformat()
        )
        self.movies.append(movie)
        self.next_id += 1
        return movie

    def find_movie(self, movie_id: int) -> Optional[Movie]:
        return next((m for m in self.movies if m.id == movie_id), None)

    def edit_movie(self, movie_id: int, **kwargs) -> bool:
        movie = self.find_movie(movie_id)
        if not movie:
            return False
        for key, value in kwargs.items():
            if hasattr(movie, key) and value is not None:
                setattr(movie, key, value)
        return True

    def delete_movie(self, movie_id: int) -> bool:
        movie = self.find_movie(movie_id)
        if movie:
            self.movies.remove(movie)
            return True
        return False

    def get_unwatched(self) -> List[Movie]:
        return [m for m in self.movies if not m.watched]

    def get_by_genre(self, genre: str) -> List[Movie]:
        return [m for m in self.movies if m.genre.lower() == genre.lower()]

    def random_movie(self, genre: Optional[str] = None) -> Optional[Movie]:
        pool = self.get_unwatched() if not genre else self.get_by_genre(genre)
        if not pool:
            # Если нет непросмотренных, возвращаем случайный из всех (или по жанру)
            pool = self.movies if not genre else self.get_by_genre(genre)
        if not pool:
            return None
        return random.choice(pool)

    def get_stats(self) -> dict:
        total = len(self.movies)
        watched = len([m for m in self.movies if m.watched])
        unwatched = total - watched
        ratings = [m.rating for m in self.movies if m.watched]
        avg_rating = sum(ratings) / len(ratings) if ratings else 0.0
        genres = {}
        for m in self.movies:
            genres[m.genre] = genres.get(m.genre, 0) + 1
        return {
            "total": total,
            "watched": watched,
            "unwatched": unwatched,
            "avg_rating": avg_rating,
            "genres": genres
        }

    def save_to_file(self, filename: str = "movies_data.json") -> None:
        data = {"movies": [asdict(m) for m in self.movies]}
        with open(filename, "w", encoding="utf-8") as f:
            json.dump(data, f, ensure_ascii=False, indent=2)

    def load_from_file(self, filename: str = "movies_data.json") -> None:
        path = Path(filename)
        if not path.exists():
            return
        with open(filename, "r", encoding="utf-8") as f:
            data = json.load(f)
            self.movies.clear()
            for item in data.get("movies", []):
                movie = Movie(
                    id=item["id"],
                    title=item["title"],
                    genre=item["genre"],
                    director=item["director"],
                    year=item["year"],
                    watched=item["watched"],
                    rating=item["rating"],
                    notes=item.get("notes", ""),
                    added_date=item["added_date"]
                )
                self.movies.append(movie)
                if movie.id >= self.next_id:
                    self.next_id = movie.id + 1

    def export_csv(self, filename: str = "movies_export.csv") -> None:
        with open(filename, "w", newline="", encoding="utf-8") as f:
            writer = csv.writer(f, delimiter=";")
            writer.writerow(["ID", "Название", "Жанр", "Режиссёр", "Год", "Просмотрен", "Оценка", "Заметки", "Дата добавления"])
            for m in self.movies:
                writer.writerow([m.id, m.title, m.genre, m.director, m.year,
                                 "Да" if m.watched else "Нет", m.rating, m.notes, m.added_date])

    def import_csv(self, filename: str = "movies_export.csv") -> None:
        path = Path(filename)
        if not path.exists():
            raise FileNotFoundError("Файл не найден")
        with open(filename, "r", encoding="utf-8") as f:
            reader = csv.DictReader(f, delimiter=";")
            for row in reader:
                try:
                    self.add_movie(
                        title=row["Название"],
                        genre=row["Жанр"],
                        director=row["Режиссёр"],
                        year=int(row["Год"]),
                        watched=row["Просмотрен"] == "Да",
                        rating=int(row["Оценка"]),
                        notes=row["Заметки"]
                    )
                except Exception as e:
                    print(f"Ошибка импорта строки: {e}")

def print_movie(movie: Movie) -> None:
    status = "✅ Просмотрен" if movie.watched else "⏳ Не просмотрен"
    print(f"#{movie.id} - {movie.title} ({movie.year})")
    print(f"   Жанр: {movie.genre}, Режиссёр: {movie.director}")
    print(f"   {status}, Оценка: {movie.rating}/10")
    if movie.notes:
        print(f"   Заметки: {movie.notes}")
    print(f"   Добавлен: {movie.added_date}")

def main():
    randomizer = MovieRandomizer()
    randomizer.load_from_file()

    while True:
        print("\n===== ГЕНЕРАТОР СЛУЧАЙНОГО ФИЛЬМА (Python) =====")
        print("1. Добавить фильм")
        print("2. Показать все фильмы")
        print("3. Рекомендовать случайный фильм")
        print("4. Рекомендовать по жанру")
        print("5. Показать непросмотренные фильмы")
        print("6. Отметить фильм как просмотренный")
        print("7. Редактировать фильм")
        print("8. Удалить фильм")
        print("9. Показать статистику")
        print("10. Сохранить в файл")
        print("11. Загрузить из файла")
        print("12. Экспорт в CSV")
        print("13. Импорт из CSV")
        print("0. Выход")
        choice = input("Выберите действие: ").strip()

        if choice == "0":
            break
        elif choice == "1":
            title = input("Название: ").strip()
            if not title:
                print("Название не может быть пустым.")
                continue
            genre = input("Жанр: ").strip()
            if not genre:
                print("Жанр не может быть пустым.")
                continue
            director = input("Режиссёр (необязательно): ").strip()
            try:
                year = int(input("Год выпуска: ").strip())
            except ValueError:
                print("Введите число.")
                continue
            watched_input = input("Статус (1-просмотрен, 0-нет): ").strip()
            watched = watched_input == "1"
            try:
                rating = int(input("Оценка (1-10): ").strip())
            except ValueError:
                rating = 0
            notes = input("Заметки (необязательно): ").strip()
            try:
                movie = randomizer.add_movie(title, genre, director, year, watched, rating, notes)
                print(f"Фильм добавлен с ID {movie.id}")
            except Exception as e:
                print("Ошибка:", e)
        elif choice == "2":
            if not randomizer.movies:
                print("Нет фильмов.")
            else:
                for m in randomizer.movies:
                    print_movie(m)
        elif choice == "3":
            movie = randomizer.random_movie()
            if not movie:
                print("Нет фильмов в коллекции.")
            else:
                print("\n🎬 Рекомендую посмотреть:")
                print_movie(movie)
        elif choice == "4":
            genre = input("Введите жанр: ").strip()
            if not genre:
                print("Введите жанр.")
                continue
            movie = randomizer.random_movie(genre)
            if not movie:
                print(f"Нет фильмов в жанре '{genre}'.")
            else:
                print(f"\n🎬 Рекомендую посмотреть в жанре {genre}:")
                print_movie(movie)
        elif choice == "5":
            unwatched = randomizer.get_unwatched()
            if not unwatched:
                print("Нет непросмотренных фильмов.")
            else:
                for m in unwatched:
                    print_movie(m)
        elif choice == "6":
            try:
                mid = int(input("Введите ID фильма: ").strip())
            except ValueError:
                print("Некорректный ID.")
                continue
            movie = randomizer.find_movie(mid)
            if not movie:
                print("Фильм не найден.")
                continue
            if movie.watched:
                print("Фильм уже отмечен как просмотренный.")
            else:
                movie.watched = True
                print("Фильм отмечен как просмотренный.")
        elif choice == "7":
            try:
                mid = int(input("Введите ID фильма для редактирования: ").strip())
            except ValueError:
                print("Некорректный ID.")
                continue
            movie = randomizer.find_movie(mid)
            if not movie:
                print("Фильм не найден.")
                continue
            print("Оставьте поле пустым, чтобы не менять.")
            new_title = input(f"Название ({movie.title}): ").strip()
            new_genre = input(f"Жанр ({movie.genre}): ").strip()
            new_director = input(f"Режиссёр ({movie.director}): ").strip()
            new_year = input(f"Год ({movie.year}): ").strip()
            new_watched = input(f"Статус (1-просмотрен, 0-нет) сейчас: {'1' if movie.watched else '0'}: ").strip()
            new_rating = input(f"Оценка ({movie.rating}): ").strip()
            new_notes = input(f"Заметки ({movie.notes}): ").strip()
            updates = {}
            if new_title: updates["title"] = new_title
            if new_genre: updates["genre"] = new_genre
            if new_director: updates["director"] = new_director
            if new_year:
                try:
                    updates["year"] = int(new_year)
                except ValueError:
                    print("Год должен быть числом, пропускаем.")
            if new_watched: updates["watched"] = new_watched == "1"
            if new_rating:
                try:
                    updates["rating"] = int(new_rating)
                except ValueError:
                    print("Оценка должна быть числом, пропускаем.")
            if new_notes: updates["notes"] = new_notes
            if randomizer.edit_movie(mid, **updates):
                print("Фильм обновлён.")
            else:
                print("Ошибка обновления.")
        elif choice == "8":
            try:
                mid = int(input("Введите ID фильма для удаления: ").strip())
            except ValueError:
                print("Некорректный ID.")
                continue
            if randomizer.delete_movie(mid):
                print("Фильм удалён.")
            else:
                print("Фильм не найден.")
        elif choice == "9":
            stats = randomizer.get_stats()
            print("\n=== СТАТИСТИКА ===")
            print(f"Всего фильмов: {stats['total']}")
            print(f"Просмотрено: {stats['watched']}")
            print(f"Не просмотрено: {stats['unwatched']}")
            print(f"Средняя оценка (просмотренные): {stats['avg_rating']:.2f}")
            print("По жанрам:")
            for g, c in stats['genres'].items():
                print(f"  {g}: {c}")
        elif choice == "10":
            randomizer.save_to_file()
            print("Сохранено.")
        elif choice == "11":
            randomizer.load_from_file()
            print("Загружено.")
        elif choice == "12":
            randomizer.export_csv()
            print("Экспортировано в movies_export.csv")
        elif choice == "13":
            try:
                randomizer.import_csv()
                print("Импортировано из movies_export.csv")
            except FileNotFoundError:
                print("Файл movies_export.csv не найден.")
            except Exception as e:
                print("Ошибка импорта:", e)
        else:
            print("Неизвестная команда.")

if __name__ == "__main__":
    main()
