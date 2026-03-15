package game

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type OpenAIProvider struct {
	APIKey   string
	BaseURL  string // Default: https://api.openai.com/v1
	Model    string // Default: gpt-4o-mini
}

func NewOpenAIProvider(apiKey, baseURL, model string) *OpenAIProvider {
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	if model == "" {
		model = "gpt-4o-mini"
	}
	return &OpenAIProvider{
		APIKey:  apiKey,
		BaseURL: baseURL,
		Model:   model,
	}
}

func (p *OpenAIProvider) Chat(ctx context.Context, systemPrompt, userMessage string, history []ChatMessage) <-chan AIResponse {
	ch := make(chan AIResponse, 1)
	go func() {
		messages := []map[string]string{
			{"role": "system", "content": systemPrompt},
		}
		for _, m := range history {
			messages = append(messages, map[string]string{"role": m.Role, "content": m.Content})
		}
		messages = append(messages, map[string]string{"role": "user", "content": userMessage})

		requestBody, _ := json.Marshal(map[string]any{
			"model":    p.Model,
			"messages": messages,
		})

		req, _ := http.NewRequestWithContext(ctx, "POST", p.BaseURL+"/chat/completions", bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+p.APIKey)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			ch <- AIResponse{Err: err}
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			ch <- AIResponse{Err: fmt.Errorf("API error (%d): %s", resp.StatusCode, string(body))}
			return
		}

		var result struct {
			Choices []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			} `json:"choices"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			ch <- AIResponse{Err: err}
			return
		}

		if len(result.Choices) > 0 {
			ch <- AIResponse{Text: result.Choices[0].Message.Content}
		} else {
			ch <- AIResponse{Err: fmt.Errorf("empty API response")}
		}
	}()
	return ch
}

func (p *OpenAIProvider) Decide(ctx context.Context, situation string, options []string) <-chan AIDecision {
	ch := make(chan AIDecision, 1)
	go func() {
		systemPrompt := "You are a game AI making a decision for an NPC. Pick one option from the provided list. Return the result in JSON format: {\"choice\": \"...\", \"reasoning\": \"...\"}"
		userMessage := fmt.Sprintf("Situation: %s\nOptions: %v", situation, options)

		// Re-using Chat internal logic or simplified version
		messages := []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userMessage},
		}

		requestBody, _ := json.Marshal(map[string]any{
			"model":           p.Model,
			"messages":        messages,
			"response_format": map[string]string{"type": "json_object"},
		})

		req, _ := http.NewRequestWithContext(ctx, "POST", p.BaseURL+"/chat/completions", bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+p.APIKey)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			ch <- AIDecision{Err: err}
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			ch <- AIDecision{Err: fmt.Errorf("API error: %d", resp.StatusCode)}
			return
		}

		var result struct {
			Choices []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			} `json:"choices"`
		}
		json.NewDecoder(resp.Body).Decode(&result)

		if len(result.Choices) > 0 {
			var dec struct {
				Choice    string `json:"choice"`
				Reasoning string `json:"reasoning"`
			}
			if err := json.Unmarshal([]byte(result.Choices[0].Message.Content), &dec); err == nil {
				ch <- AIDecision{ChosenOption: dec.Choice, Reasoning: dec.Reasoning}
			} else {
				ch <- AIDecision{Err: err}
			}
		} else {
			ch <- AIDecision{Err: fmt.Errorf("empty API response")}
		}
	}()
	return ch
}
