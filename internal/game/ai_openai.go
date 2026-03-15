package game

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
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
		BaseURL: strings.TrimSuffix(baseURL, "/"),
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

		DebugLog("AI Request: POST %s/chat/completions", p.BaseURL)
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
			DebugLog("[AI-ERROR] POST %s/chat/completions failed with status: %d", p.BaseURL, resp.StatusCode)
			if resp.StatusCode == 429 {
				DebugLog("[AI-ERROR] Rate limited by provider (429).")
			}
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
			text := result.Choices[0].Message.Content
			DebugLog("[AI-VERBOSE] Provider Response: %s", text)
			ch <- AIResponse{Text: text}
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

		DebugLog("AI Request: POST %s/chat/completions", p.BaseURL)
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
			body, _ := io.ReadAll(resp.Body)
			DebugLog("[AI-ERROR] POST %s/chat/completions failed with status: %d. Body: %s", p.BaseURL, resp.StatusCode, string(body))
			if resp.StatusCode == 429 {
				DebugLog("[AI-ERROR] Rate limited by provider (429).")
			}
			ch <- AIDecision{Err: fmt.Errorf("API error (%d): %s", resp.StatusCode, string(body))}
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
			content := result.Choices[0].Message.Content
			DebugLog("[AI-VERBOSE] Provider Decision JSON: %s", content)
			var dec struct {
				Choice    string `json:"choice"`
				Reasoning string `json:"reasoning"`
			}
			if err := json.Unmarshal([]byte(content), &dec); err == nil {
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
func (p *OpenAIProvider) ListModels(ctx context.Context) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", p.BaseURL+"/models", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+p.APIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var models []string
	for _, m := range result.Data {
		models = append(models, m.ID)
	}
	return models, nil
}
