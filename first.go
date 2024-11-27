// CheckActivationStatus check sims activation status
func (w *Workflow) CheckActivationStatus(ctx context.Context, in entity.AsrCheckSimsStatuses) (entity.AsrSimsStatuses, error) {
	log := observability.GetLogger(ctx)
	log.InfoContext(ctx, "start CheckActivationStatus", "params", in)

	ticker := time.NewTicker(w.CheckAsrStatusInterval)

	activationStatuses := &entity.AsrSimsStatuses{}
	var err error

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		for {
			select {
			case _ = <-ticker.C:
				activationStatuses, err = w.AsrAdapterService.CheckSimsStatuses(ctx, &in)
				if err != nil {
					log.ErrorContext(ctx, "can`t get asr activation status", "requestId", in.AsrRequestID)

					ticker.Stop()
					wg.Done()

					return
				}

				log.InfoContext(ctx, "asr activation", "status", activationStatuses.CommandStatus)

				if activationStatuses.CommandStatus != asrCompletedActivationStatus &&
					activationStatuses.CommandStatus != asrFaultedActivationStatus &&
					activationStatuses.CommandStatus != asrCompletedPartialActivationStatus {
					continue
				}

				ticker.Stop()
				wg.Done()

				return
			}
		}
	}()

	wg.Wait()

	if err != nil {
		return entity.AsrSimsStatuses{}, err
	}

	return *activationStatuses, nil
}
