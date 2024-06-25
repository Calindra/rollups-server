package container

import (
	"github.com/calindra/rollups-server/src/decoder"
	"github.com/calindra/rollups-server/src/model"
	"github.com/calindra/rollups-server/src/services"
	"github.com/jmoiron/sqlx"
)

type Container struct {
	db                 *sqlx.DB
	outputDecoder      *decoder.OutputDecoder
	convenienceService *services.ConvenienceService
	repository         *model.VoucherRepository
	noticeRepository   *model.NoticeRepository
}

func NewContainer(db sqlx.DB) *Container {
	return &Container{
		db: &db,
	}
}

func (c *Container) GetOutputDecoder() *decoder.OutputDecoder {
	if c.outputDecoder != nil {
		return c.outputDecoder
	}
	c.outputDecoder = decoder.NewOutputDecoder(*c.GetConvenienceService())
	return c.outputDecoder
}

func (c *Container) GetConvenienceService() *services.ConvenienceService {
	if c.convenienceService != nil {
		return c.convenienceService
	}
	c.convenienceService = services.NewConvenienceService(
		c.GetRepository(),
		c.GetNoticeRepository(),
	)
	return c.convenienceService
}

func (c *Container) GetRepository() *model.VoucherRepository {
	if c.repository != nil {
		return c.repository
	}
	c.repository = &model.VoucherRepository{
		Db: *c.db,
	}
	err := c.repository.CreateTables()
	if err != nil {
		panic(err)
	}
	return c.repository
}

func (c *Container) GetNoticeRepository() *model.NoticeRepository {
	if c.noticeRepository != nil {
		return c.noticeRepository
	}
	c.noticeRepository = &model.NoticeRepository{
		Db: *c.db,
	}
	err := c.noticeRepository.CreateTables()
	if err != nil {
		panic(err)
	}
	return c.noticeRepository
}
