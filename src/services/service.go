package services

import (
	"context"

	"github.com/calindra/rollups-server/src/model"
	"github.com/calindra/rollups-server/src/util"
)

type ConvenienceService struct {
	voucherRepository *model.VoucherRepository
	noticeRepository  *model.NoticeRepository
}

func NewConvenienceService(
	voucherRepository *model.VoucherRepository,
	noticeRepository *model.NoticeRepository,
) *ConvenienceService {
	return &ConvenienceService{
		voucherRepository: voucherRepository,
		noticeRepository:  noticeRepository,
	}
}

func (s *ConvenienceService) CreateVoucher1(
	ctx context.Context,
	voucher *model.ConvenienceVoucher,
) (*model.ConvenienceVoucher, error) {
	return s.voucherRepository.CreateVoucher(ctx, voucher)
}

func (s *ConvenienceService) CreateNotice(
	ctx context.Context,
	notice *model.ConvenienceNotice,
) (*model.ConvenienceNotice, error) {
	noticeInDb, err := s.noticeRepository.FindByInputAndOutputIndex(
		ctx, notice.InputIndex, notice.OutputIndex,
	)
	if err != nil {
		return nil, err
	}

	if noticeInDb != nil {
		return s.noticeRepository.Update(ctx, notice)
	}
	return s.noticeRepository.Create(ctx, notice)
}
func (s *ConvenienceService) CreateVoucher(
	ctx context.Context,
	voucher *model.ConvenienceVoucher,
) (*model.ConvenienceVoucher, error) {

	voucherInDb, err := s.voucherRepository.FindVoucherByInputAndOutputIndex(
		ctx, voucher.InputIndex,
		voucher.OutputIndex,
	)

	if err != nil {
		return nil, err
	}

	if voucherInDb != nil {
		return s.voucherRepository.UpdateVoucher(ctx, voucher)
	}

	return s.voucherRepository.CreateVoucher(ctx, voucher)
}

func (c *ConvenienceService) UpdateExecuted(
	ctx context.Context,
	inputIndex uint64,
	outputIndex uint64,
	executedValue bool,
) error {
	return c.voucherRepository.UpdateExecuted(
		ctx,
		inputIndex,
		outputIndex,
		executedValue,
	)
}

func (c *ConvenienceService) FindAllVouchers(
	ctx context.Context,
	first *int,
	last *int,
	after *string,
	before *string,
	filter []*model.ConvenienceFilter,
) (*util.PageResult[model.ConvenienceVoucher], error) {
	return c.voucherRepository.FindAllVouchers(
		ctx,
		first,
		last,
		after,
		before,
		filter,
	)
}

func (c *ConvenienceService) FindAllNotices(
	ctx context.Context,
	first *int,
	last *int,
	after *string,
	before *string,
	filter []*model.ConvenienceFilter,
) (*util.PageResult[model.ConvenienceNotice], error) {
	return c.noticeRepository.FindAllNotices(
		ctx,
		first,
		last,
		after,
		before,
		filter,
	)
}

func (c *ConvenienceService) FindVoucherByInputAndOutputIndex(
	ctx context.Context, inputIndex uint64, outputIndex uint64,
) (*model.ConvenienceVoucher, error) {
	return c.voucherRepository.FindVoucherByInputAndOutputIndex(
		ctx, inputIndex, outputIndex,
	)
}

func (c *ConvenienceService) FindNoticeByInputAndOutputIndex(
	ctx context.Context, inputIndex uint64, outputIndex uint64,
) (*model.ConvenienceNotice, error) {
	return c.noticeRepository.FindByInputAndOutputIndex(
		ctx, inputIndex, outputIndex,
	)
}
