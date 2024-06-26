package movies

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"project/internal/database/movie"
	"project/internal/utils"

	"github.com/gin-gonic/gin"
)

// @Tags			movies
// @Summary		Create a new movie
// @Description	Creates a new movie with provided details
// @Produce		json
// @Security		ApiKeyAuth
// @Param			movie	body		routes.MovieCreateRequest		true	"Movie object"
// @Success		200		{object}	routes.DefaultMessageResponse	"Movie Created"
// @Failure		400		{object}	routes.DefaultMessageResponse	"Bad request"
// @Failure		401		{object}	routes.DefaultMessageResponse	"Unauthorized"
// @Failure		404		{object}	routes.DefaultMessageResponse	"Genre not found"
// @Failure		404		{object}	routes.DefaultMessageResponse	"Age category not found"
// @Failure		404		{object}	routes.DefaultMessageResponse	"Category not found"
// @Failure		500		{object}	routes.DefaultMessageResponse	"Internal server error"
// @Router			/api/movies [post]
func (m *MoviesRoute) POST_Movie(c *gin.Context) {
	userRole := c.GetString("role")
	userId := c.GetInt("userId")

	if userRole != "admin" || userId == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "Unauthorized",
		})
		return
	}
	var movie movie.Movie

	err := c.BindJSON(&movie)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Bad request",
		})
		return
	}

	if len(movie.Genres) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "At least one genre required",
		})
		return
	}

	exists, err := m.DB.AgeRepository.CheckAgeCategoryExistsId(movie.AgeCategoryId)
	if err != nil {
		log.Println(err.Error())
		c.JSON(500, gin.H{
			"message": "Internal server error",
		})
		return
	}
	if !exists {
		c.JSON(400, gin.H{
			"message": "This age category does not exists",
		})
		return
	}
	exists, err = m.DB.CategoriesRepository.CheckCategoryExistsById(movie.CategoryId)
	if err != nil {
		log.Println(err.Error())
		c.JSON(500, gin.H{
			"message": "Internal server error",
		})
		return
	}
	if !exists {
		c.JSON(400, gin.H{
			"message": "This category does not exists",
		})
		return
	}
	genreIds := []int{}
	for _, genre := range movie.Genres {
		id, _, err := m.DB.GenreRepository.CheckGenreExistsByName(genre)
		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(404, gin.H{
					"message": "This genre (" + genre + ") does not exists",
				})
				return

			}
			log.Println(err.Error())
			c.JSON(500, gin.H{
				"message": "Internal server error",
			})
			return
		}
		genreIds = append(genreIds, id)
	}
	resultMovie, err := m.DB.MovieRepository.CreateMovie(
		userId,
		movie.Name,
		movie.Year,
		movie.CategoryId,
		movie.AgeCategoryId,
		movie.DurationMinutes,
		movie.Description,
		movie.Keywords,
		movie.Director,
		movie.Producer)
	if err != nil {
		log.Println(err.Error())
		c.JSON(500, gin.H{
			"message": "Internal server error",
		})
		return
	}
	err = m.DB.MovieRepository.AddGenresToMovie(resultMovie.Id, genreIds)
	if err != nil {
		log.Println(err.Error())
		c.JSON(500, gin.H{
			"message": "Internal server error",
		})
		return
	}
	c.JSON(200, gin.H{
		"message": "Movie Created",
	})
}

// +1 watch by authorised user
// @Summary Watch a movie
// @Description Records a user watching a movie
// @Tags movies
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "Movie ID"
// @Success 200 {object} routes.DefaultMessageResponse "Successful operation"
// @Failure 400 {object} routes.DefaultMessageResponse "Bad request"
// @Failure 401 {object} routes.DefaultMessageResponse "Unauthorized"
// @Failure 404 {object} routes.DefaultMessageResponse "Movie not found"
// @Failure 500 {object} routes.DefaultMessageResponse "Internal server error"
// @Router /api/movies/watch/{id} [post]
func (m *MoviesRoute) POST_Watch(c *gin.Context) {
	userId := c.GetInt("userId")
	movieId := c.Param("id")
	valid, movieIdNum := utils.IsValidNum(movieId)
	if !valid {
		c.JSON(400, gin.H{
			"message": "Bad request",
		})
		return
	}
	if userId == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "Unauthorized",
		})
		return
	}

	_, err := m.DB.MovieRepository.GetMovieById(userId, movieIdNum)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(404, gin.H{
				"message": "Movie not found",
			})
			return
		}
		log.Println(err.Error())
		c.JSON(500, gin.H{
			"message": "Internal server error",
		})
		return
	}
	count, err := m.DB.MovieRepository.MovieWasWatchedByUser(movieIdNum)
	if err != nil {
		log.Println(err.Error())
		c.JSON(500, gin.H{
			"message": "Internal server error",
		})
		return
	}
	message := fmt.Sprintf("Total watches: %d", count)
	c.JSON(200, gin.H{
		"message": message,
	})

}
