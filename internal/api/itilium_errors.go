package api

import "fmt"

// ResponsibleChangeNotConfirmedError means ITILIUM returned success HTTP but find_sc still shows the previous assignee.
type ResponsibleChangeNotConfirmedError struct {
	Number                 string
	RequestedResponsibleID string
	ActualResponsibleID    string
	ActualResponsibleTitle string
}

func (e *ResponsibleChangeNotConfirmedError) Error() string {
	if e == nil {
		return "itilium did not confirm responsible change"
	}
	if title := e.ActualResponsibleTitle; title != "" {
		return fmt.Sprintf(
			"itilium did not confirm responsible change for ticket %s (still %s)",
			e.Number,
			title,
		)
	}
	return fmt.Sprintf("itilium did not confirm responsible change for ticket %s", e.Number)
}
