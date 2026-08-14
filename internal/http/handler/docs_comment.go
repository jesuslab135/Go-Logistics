package handler

// docCreateComment godoc
//
//	@Summary	Create comments
//	@Tags	comments
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	body	body	dto.CreateCommentRequest	true	"body"
//	@Success	201	{object}	dto.CommentResponse
//	@Failure	400	{object}	dto.ErrorResponse
//	@Router	/api/v1/comments [post]
func docCreateComment() {}

// docGetComment godoc
//
//	@Summary	Get comments
//	@Tags	comments
//	@Security	BearerAuth
//	@Produce	json
//	@Param	id	path	int	true	"id"
//	@Success	200	{object}	dto.CommentResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/comments/{id} [get]
func docGetComment() {}

// docUpdateComment godoc
//
//	@Summary	Update comments
//	@Tags	comments
//	@Security	BearerAuth
//	@Accept	json
//	@Produce	json
//	@Param	id	path	int	true	"id"
//	@Param	body	body	dto.UpdateCommentRequest	true	"body"
//	@Success	200	{object}	dto.CommentResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router	/api/v1/comments/{id} [put]
func docUpdateComment() {}

// docDeleteComment godoc
//
//	@Summary	Delete comments
//	@Tags	comments
//	@Security	BearerAuth
//	@Param	id	path	int	true	"id"
//	@Success	204	"No Content"
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router	/api/v1/comments/{id} [delete]
func docDeleteComment() {}
