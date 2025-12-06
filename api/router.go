package api

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/jackc/pgx/v5/pgtype"

	"golang-api/db"
)

type SubscriptionListOutput struct {
	Body struct {
		Subscriptions []db.Subscription `json:"subscriptions" doc:"List of subscriptions"`
	}
}

type SubscriptionOutput struct {
	Body struct {
		Subscription db.Subscription `json:"subscription" doc:"Subscription details"`
	}
}

type CreateSubscriptionInput struct {
	Body struct {
		ServiceName string  `json:"service_name" example:"Yandex Plus" doc:"Service Name"`
		Price       int     `json:"price" example:"400" doc:"Price"`
		UserID      string  `json:"user_id" example:"123e4567-e89b-12d3-a456-426614174000" doc:"User ID"`
		StartDate   string  `json:"start_date" example:"01-2025" doc:"Start Date"`
		EndDate     *string `json:"end_date,omitempty" example:"07-2025" doc:"End Date"`
	} `json:"body"`
}

func SetupRouter(db *db.Queries) *http.ServeMux {
	routerMux := http.NewServeMux()

	api := humago.New(routerMux, huma.DefaultConfig("Golang API", "1.0.0"))

	registerRoutes(api, db)

	// Add a custom 404 handler for non-existent routes
	routerMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Check if the request was handled by Huma routes
		if r.URL.Path != "/" {
			w.Header().Set("Content-Type", "application/problem+json")
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, `{"title": "Not Found", "status": 404, "detail": "Route not found: %s"}`, r.URL.Path)
			return
		}
		
		// Handle the root path with a welcome message
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"message": "Welcome to the API", "documentation": "/docs"}`)
	})

	return routerMux
}

func registerRoutes(api huma.API, queries *db.Queries) {
	//GET list-subscription
	huma.Register(api, huma.Operation{
		OperationID: "list-subscription",
		Method:      http.MethodGet,
		Path:        "/api/subscriptions/list",
		Summary:     "Get a list of subscriptions",
		Description: "Get a list of subscriptions.",
		Tags:        []string{"Subscriptions"},
	}, func(ctx context.Context, input *struct{}) (*SubscriptionListOutput, error) {
		subscriptions, err := queries.ListSubscriptions(ctx)
		if err != nil {
			slog.Error("Failed to list subscriptions", "error", err)
			return nil, err
		}

		resp := &SubscriptionListOutput{}
		resp.Body.Subscriptions = subscriptions
		return resp, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "view-subscription",
		Method:      http.MethodGet,
		Path:        "/api/subscriptions/{id}/view",
		Summary:     "Get a subscription by ID",
		Description: "Get a subscription by ID.",
		Tags:        []string{"Subscriptions"},
	}, func(ctx context.Context, input *struct {
		ID int32 `path:"id"`
	}) (*SubscriptionOutput, error) {
		subscription, err := queries.GetSubscription(ctx, input.ID)
		if err != nil {
			slog.Error("Failed to get subscription by ID", "id", input.ID, "error", err)
			return nil, &huma.ErrorModel{
				Title:  "Bad Request",
				Status: http.StatusBadRequest,
				Detail: "Failed to get subscription by ID",
			}
		}

		resp := &SubscriptionOutput{}
		resp.Body.Subscription = subscription
		return resp, nil
	})

	huma.Register(api, huma.Operation{
		OperationID:   "create-subscription",
		Method:        http.MethodPost,
		Path:          "/api/subscriptions/create",
		Summary:       "Create a new subscription",
		Description:   "Create a new subscription",
		DefaultStatus: http.StatusCreated,
		Tags:          []string{"Subscriptions"},
	}, func(ctx context.Context, input *CreateSubscriptionInput) (*SubscriptionOutput, error) {
		// Parse dates
		t, err := time.Parse("01-2006", input.Body.StartDate)
		if err != nil {
			slog.Error("Error parsing start date: ", err)
			return nil, err
		}

		startDate := pgtype.Timestamp{
			Time:  t,
			Valid: true,
		}

		var endDate pgtype.Timestamp
		if input.Body.EndDate != nil {
			t, err = time.Parse("01-2006", *input.Body.EndDate)
			if err != nil {
				slog.Error("Error parsing end date: ", err)
				return nil, err
			}

			endDate = pgtype.Timestamp{
				Time:  t,
				Valid: true,
			}
		} else {
			endDate = pgtype.Timestamp{}
		}

		params := db.CreateSubscriptionParams{
			ServiceName: input.Body.ServiceName,
			Price:       int32(input.Body.Price),
			UserID:      input.Body.UserID,
			StartDate:   startDate,
			EndDate:     endDate,
		}

		slog.Info("Creating subscription", "params", params)

		createdSubscription, err := queries.CreateSubscription(ctx, params)
		if err != nil {
			slog.Error("Failed to create subscription", "error", err)
			return nil, err
		}

		resp := &SubscriptionOutput{}
		resp.Body.Subscription = createdSubscription
		return resp, nil
	})

	type UpdateSubscriptionInput struct {
		ID   int32 `path:"id"`
		Body struct {
			ServiceName *string `json:"service_name,omitempty" example:"Yandex Plus" doc:"Service Name"`
			Price       *int    `json:"price,omitempty" example:"400" doc:"Price"`
			UserID      *string `json:"user_id,omitempty" example:"123e4567-e89b-12d3-a456-426614174000" doc:"User ID"`
			StartDate   *string `json:"start_date,omitempty" example:"01-2025" doc:"Start Date"`
			EndDate     *string `json:"end_date,omitempty" example:"07-2025" doc:"End Date"`
		} `json:"body"`
	}

	huma.Register(api, huma.Operation{
		OperationID:   "update-subscription",
		Method:        http.MethodPost,
		Path:          "/api/subscriptions/{id}/update",
		Summary:       "Update a subscription",
		Description:   "Update a subscription",
		DefaultStatus: http.StatusOK,
		Tags:          []string{"Subscriptions"},
	}, func(ctx context.Context, input *UpdateSubscriptionInput) (*SubscriptionOutput, error) {
		// Prepare parameters with optional fields
		params := db.UpdateSubscriptionParams{
			ID: input.ID,
		}

		// Handle UserID
		if input.Body.UserID != nil {
			params.Column2 = *input.Body.UserID
		} else {
			params.Column2 = ""
		}

		// Handle ServiceName
		if input.Body.ServiceName != nil {
			params.Column3 = *input.Body.ServiceName
		} else {
			params.Column3 = ""
		}

		// Handle Price
		if input.Body.Price != nil {
			price := int32(*input.Body.Price)
			params.Column4 = price
		} else {
			params.Column4 = 0
		}

		// Handle StartDate
		if input.Body.StartDate != nil {
			t, err := time.Parse("01-2006", *input.Body.StartDate)
			if err != nil {
				slog.Error("Error parsing start date: ", err)
				return nil, err
			}
			startDate := pgtype.Timestamp{
				Time:  t,
				Valid: true,
			}
			params.Column5 = startDate
		} else {
			params.Column5 = pgtype.Timestamp{}
		}

		// Handle EndDate
		if input.Body.EndDate != nil {
			t, err := time.Parse("01-2006", *input.Body.EndDate)
			if err != nil {
				slog.Error("Error parsing end date: ", err)
				return nil, err
			}
			endDate := pgtype.Timestamp{
				Time:  t,
				Valid: true,
			}
			params.Column6 = endDate
		} else {
			params.Column6 = pgtype.Timestamp{}
		}

		err := queries.UpdateSubscription(ctx, params)
		if err != nil {
			slog.Error("Failed to update subscription", "error", err)
			return nil, err
		}


		subscription, err := queries.GetSubscription(ctx, input.ID)
		if err != nil {
			slog.Error("Failed to get subscription by ID", "id", input.ID, "error", err)
			return nil, err
		}

		resp := &SubscriptionOutput{}
		resp.Body.Subscription = subscription
		return resp, nil
	})

	huma.Register(api, huma.Operation{
		OperationID:   "delete-subscription",
		Method:        http.MethodPost,
		Path:          "/api/subscriptions/{id}/delete",
		Summary:       "Delete subscription",
		Description:   "Delete subscription",
		DefaultStatus: http.StatusOK,
		Tags:          []string{"Subscriptions"},
	}, func(ctx context.Context, input *struct {
		ID int32 `path:"id"`
	}) (*struct {
		Message string `json:"message"`
	}, error) {
		err := queries.DeleteSubscription(ctx, input.ID)
		if err != nil {
			slog.Error("Failed to delete subscription", "error", err)
			return nil, err
		}

		return &struct {
			Message string `json:"message"`
		}{Message: "Subscription deleted successfully"}, nil
	})

	type GetTotalValueSubscriptionInput struct {
		Body struct {
			ServiceName string  `json:"service_name" example:"Yandex Plus" doc:"Service Name"`
			UserID      string  `json:"user_id" example:"123e4567-e89b-12d3-a456-426614174000" doc:"User ID"`
			StartDate   string  `json:"start_date,omitempty" example:"01-2025" doc:"Start Date"`
			EndDate     *string `json:"end_date,omitempty" example:"07-2025" doc:"End Date"`
		}
	}

	type GetTotalValueSubscriptionOutput struct {
		Body struct {
			TotalValue int64 `json:"total_value" doc:"Total Value"`
		}
	}

	huma.Register(api, huma.Operation{
		OperationID: "get-total-value-subscription",
		Method:      http.MethodGet,
		Path:        "/api/subscriptions/total-value",
		Summary:     "Get the total value of subscriptions",
		Description: "Get the total value of subscriptions.",
		Tags:        []string{"Subscriptions"},
	}, func(ctx context.Context, input *GetTotalValueSubscriptionInput) (*GetTotalValueSubscriptionOutput, error) {
		var startDate pgtype.Timestamp
		if input.Body.StartDate != "" {
			t, err := time.Parse("01-2006", input.Body.StartDate)
			if err != nil {
				slog.Error("Error parsing start date: ", err)
				return nil, err
			}

			startDate = pgtype.Timestamp{
				Time:  t,
				Valid: true,
			}
		} else {
			// If no start date is provided, get the earliest start date for this user
			earliestDateResult, err := queries.GetEarliestStartDateForUser(ctx, input.Body.UserID)
			if err != nil {
				slog.Error("Error getting earliest start date for user: ", err)
				return nil, err
			}
			
			// Convert the result to pgtype.Timestamp
			var earliestDate pgtype.Timestamp
			if earliestDateResult != nil {
				if t, ok := earliestDateResult.(time.Time); ok {
					earliestDate = pgtype.Timestamp{
						Time:  t,
						Valid: true,
					}
				}
			}
			startDate = earliestDate
		}

		var endDate pgtype.Timestamp
		if input.Body.EndDate != nil {
			t, err := time.Parse("01-2006", *input.Body.EndDate)
			if err != nil {
				slog.Error("Error parsing end date: ", err)
				return nil, err
			}

			endDate = pgtype.Timestamp{
				Time:  t,
				Valid: true,
			}
		} else {
			// If no end date is provided, use current date
			endDate = pgtype.Timestamp{
				Time:  time.Now(),
				Valid: true,
			}
		}

		params := db.GetTotalValueSubscriptionParams{
			UserID:      input.Body.UserID,
			ServiceName: input.Body.ServiceName,
			StartDate:   startDate,
			StartDate_2: endDate,
			Column5:     true,  // Filter by user ID
			Column6:     input.Body.ServiceName != "", // Filter by service name only if provided
		}

		totalValueResult, err := queries.GetTotalValueSubscription(ctx, params)
		if err != nil {
			slog.Error("Failed to get total value of subscriptions", "error", err)
			return nil, err
		}

		// Convert the result to int64
		var totalValue int64
		if totalValueResult != nil {
			switch v := totalValueResult.(type) {
			case int64:
				totalValue = v
			case int32:
				totalValue = int64(v)
			case int:
				totalValue = int64(v)
			default:
				totalValue = 0
			}
		}

		resp := &GetTotalValueSubscriptionOutput{}
		resp.Body.TotalValue = totalValue
		return resp, nil
	})
}
