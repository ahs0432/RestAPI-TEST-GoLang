package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

// 유저 정보 저장 JSON 형식
type User struct {
	NickName string `json:"nickname"`
	Email    string `json:"email"`
	Etc      string `json:"Etc"`
}

// 오류 코드 관리를 위한 JSON
type ErrorList struct {
	ErrorCode int    `json:"errorCode"`
	Message   string `json:"message"`
}

// User 임시 데이터
var users = map[int]*User{}
var lastIndex = 1

func main() {
	router := httprouter.New()

	// /user 페이지 GET 접근 시 전체 호출 확인
	usersGetHandle := httprouter.Handle(func(writer http.ResponseWriter, req *http.Request, param httprouter.Params) {
		writer.Header().Add("Content-Type", "application/json")

		// 기본 Limit 개수와 offset을 기본 값으로 설정해둠
		limit := 25
		offset := 1

		// URL 내 Query 값이 빈 값이 아닌 경우
		if req.URL.Query() != nil {
			reqQuery := req.URL.Query()

			// Limit Query를 확인하고 Integer 값이라면 1~25 내의 값인지 확인 후 반영
			_, exist := reqQuery["limit"]
			if exist {
				lmt, err := strconv.Atoi(reqQuery["limit"][0])
				if err != nil {
					writer.WriteHeader(http.StatusBadRequest)
					errorCode := ErrorList{ErrorCode: 400, Message: "Limit Query is not Integer."}
					json.NewEncoder(writer).Encode(errorCode)
				} else {
					if lmt > 0 && lmt <= 25 {
						limit = lmt
					} else {
						writer.WriteHeader(http.StatusBadRequest)
						errorCode := ErrorList{ErrorCode: 400, Message: "Limit Query value range is from 1 to 25."}
						json.NewEncoder(writer).Encode(errorCode)
						return
					}
				}
			}

			// Offset Query를 확인하고 Integer 값이라면 1~[lastindex] 내의 값인지 확인 후 반영
			_, exist = reqQuery["offset"]
			if exist {
				ofs, err := strconv.Atoi(reqQuery["offset"][0])
				if err != nil {
					writer.WriteHeader(http.StatusBadRequest)
					errorCode := ErrorList{ErrorCode: 400, Message: "Offset Query is not Integer."}
					json.NewEncoder(writer).Encode(errorCode)
				} else {
					if ofs > 0 && ofs < lastIndex {
						offset = ofs

					} else {
						writer.WriteHeader(http.StatusBadRequest)
						errorCode := ErrorList{ErrorCode: 400, Message: "Offset Query value range is from ~" + strconv.Itoa(lastIndex-1)}
						json.NewEncoder(writer).Encode(errorCode)
						return
					}
				}
			}
		}

		// offset + limit 값이 최대 값보다 큰 경우 불필요한 오버헤드를 방지하기 위해 계산.
		if (offset + limit - 1) > lastIndex {
			limit -= (lastIndex - (offset + limit))
		}

		// 현재 조회된 User 값을 보관할 Map을 하나 생성하고 For 문을 통해 대상 값의 유무 확인 후 Map 내에 담아 전달
		user := map[int]*User{}

		for i := offset; i < (offset + limit); i++ {
			u, exist := users[i]
			if exist {
				user[i] = u
			}
		}

		json.NewEncoder(writer).Encode(user)
	})

	// /user 페이지 POST 접근 시 Last Index로 추가
	userPostHandle := httprouter.Handle(func(writer http.ResponseWriter, req *http.Request, param httprouter.Params) {
		writer.Header().Add("Content-Type", "application/json")

		var user User
		// HTTP 요청을 수신받아 Decode 하여 User Struct에 변수로 입력
		json.NewDecoder(req.Body).Decode(&user)

		// 필수 컨텐츠 존재 여부 확인 후 미존재 시 400 에러 발생
		if user.Email == "" {
			writer.WriteHeader(http.StatusBadRequest)
			errorCode := ErrorList{ErrorCode: 400, Message: "Required element(Email) is null."}
			json.NewEncoder(writer).Encode(errorCode)
		} else if user.NickName == "" {
			writer.WriteHeader(http.StatusBadRequest)
			errorCode := ErrorList{ErrorCode: 400, Message: "Required element(NickName) is null."}
			json.NewEncoder(writer).Encode(errorCode)
		} else {
			// 문제 없을 경우 현재 Index에 값 추가 후 인덱스 값을 늘림
			users[lastIndex] = &user
			lastIndex++

			writer.WriteHeader(http.StatusCreated)
			json.NewEncoder(writer).Encode(user)
		}
	})

	// /user/[index] 페이지 GET 접근 시 대상 Index 유무 확인 및 사용자에게 데이터 전송
	userGetHandle := httprouter.Handle(func(writer http.ResponseWriter, req *http.Request, param httprouter.Params) {
		writer.Header().Add("Content-Type", "application/json")
		index, err := strconv.Atoi(param.ByName("idx"))

		// [index] 값이 Integer 값이 맞는지 확인 후 아닐 경우 400 에러 발생
		if err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			errorCode := ErrorList{ErrorCode: 400, Message: param.ByName("idx") + "(Index) is not Integer."}
			json.NewEncoder(writer).Encode(errorCode)
		} else {
			user, exists := users[index]

			// [index] 값이 존재하는 값인지 여부를 확인하고 미존재 시 404 에러 발생 / 정상일 경우 사용자에게 출력
			if exists {
				json.NewEncoder(writer).Encode(user)
			} else {
				writer.WriteHeader(http.StatusNotFound)
				errorCode := ErrorList{ErrorCode: 404, Message: param.ByName("idx") + "(Index) is not Found."}
				json.NewEncoder(writer).Encode(errorCode)
			}
		}
	})

	// /user/[index] 페이지 PUT 접근 시 대상 Index 유무 확인 및 사용자 데이터 수정 후 수정된 데이터 전달
	userPutHandle := httprouter.Handle(func(writer http.ResponseWriter, req *http.Request, param httprouter.Params) {
		writer.Header().Add("Content-Type", "application/json")
		index, err := strconv.Atoi(param.ByName("idx"))

		// [index] 값이 Integer 값이 맞는지 확인 후 아닐 경우 400 에러 발생
		if err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			errorCode := ErrorList{ErrorCode: 400, Message: param.ByName("idx") + "(Index) is not Integer."}
			json.NewEncoder(writer).Encode(errorCode)
		} else {
			_, exists := users[index]

			// [index] 값이 존재하는 값인지 여부를 확인하고 미존재 시 404 에러 발생
			if exists {
				var user User
				json.NewDecoder(req.Body).Decode(&user)

				// 필수 값의 경우 비워두면 기본 값이 없기에 오류 발생 (Email, NickName) / 문제 없을 시 이외 값은 기본 값으로 지정하여 변경 진행
				if user.Email == "" {
					writer.WriteHeader(http.StatusBadRequest)
					errorCode := ErrorList{ErrorCode: 400, Message: "Required element(Email) is null."}
					json.NewEncoder(writer).Encode(errorCode)
				} else if user.NickName == "" {
					writer.WriteHeader(http.StatusBadRequest)
					errorCode := ErrorList{ErrorCode: 400, Message: "Required element(NickName) is null."}
					json.NewEncoder(writer).Encode(errorCode)
				} else {
					users[index] = &user
					json.NewEncoder(writer).Encode(users[index])
				}
			} else {
				writer.WriteHeader(http.StatusNotFound)
				errorCode := ErrorList{ErrorCode: 404, Message: param.ByName("idx") + "(Index) is not Found."}
				json.NewEncoder(writer).Encode(errorCode)
			}
		}
	})

	// /user/[index] 페이지 PUT 접근 시 대상 Index 유무 확인 및 사용자 데이터 삭제 후 삭제된 데이터 전달
	userDeleteHandle := httprouter.Handle(func(writer http.ResponseWriter, req *http.Request, param httprouter.Params) {
		writer.Header().Add("Content-Type", "application/json")
		index, err := strconv.Atoi(param.ByName("idx"))

		// [index] 값이 Integer 값이 맞는지 확인 후 아닐 경우 400 에러 발생
		if err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			errorCode := ErrorList{ErrorCode: 400, Message: param.ByName("idx") + "(Index) is not Integer."}
			json.NewEncoder(writer).Encode(errorCode)
		} else {
			user, exists := users[index]

			// [index] 값이 존재하는 값인지 여부를 확인하고 미존재 시 404 에러 발생 / 정상일 경우 데이터 삭제 후 삭제한 데이터 출력
			if exists {
				delete(users, index)
				json.NewEncoder(writer).Encode(user)
			} else {
				writer.WriteHeader(http.StatusNotFound)
				errorCode := ErrorList{ErrorCode: 404, Message: param.ByName("idx") + "(Index) is not Found."}
				json.NewEncoder(writer).Encode(errorCode)
			}
		}
	})

	// /user/[index] 페이지 PATCH 접근 시 대상 Index 유무 확인 및 수정 요청 값에 대해서만 수정 후 수정된 데이터 전달
	userPatchHandle := httprouter.Handle(func(writer http.ResponseWriter, req *http.Request, param httprouter.Params) {
		writer.Header().Add("Content-Type", "application/json")
		index, err := strconv.Atoi(param.ByName("idx"))

		// [index] 값이 Integer 값이 맞는지 확인 후 아닐 경우 400 에러 발생
		if err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			errorCode := ErrorList{ErrorCode: 400, Message: param.ByName("idx") + "(Index) is not Integer."}
			json.NewEncoder(writer).Encode(errorCode)
		} else {
			_, exists := users[index]

			// [index] 값이 존재하는 값인지 여부를 확인하고 미존재 시 404 에러 발생 / 정상일 경우 데이터 삭제 후 삭제한 데이터 출력
			if exists {
				var user User
				json.NewDecoder(req.Body).Decode(&user)

				// 변경된 것이 있는지 확인하는 변수
				changeCheck := false

				// Email, NickName, Etc 중 변경 요청 값만 변경하고 변경됐다면 changeCheck 변수에 반영
				if user.Email != "" {
					users[index].Email = user.Email
					changeCheck = true
				}

				if user.NickName != "" {
					users[index].NickName = user.NickName
					changeCheck = true
				}

				if user.Etc != "" {
					users[index].Etc = user.Etc
					changeCheck = true
				}

				// 하나의 값도 변경되지 않았다면 400 에러 발생 / 정상인 경우 변경 이후 데이터 출력
				if changeCheck {
					json.NewEncoder(writer).Encode(users[index])
				} else {
					writer.WriteHeader(http.StatusBadRequest)
					errorCode := ErrorList{ErrorCode: 400, Message: "All element is null."}
					json.NewEncoder(writer).Encode(errorCode)
				}
			} else {
				writer.WriteHeader(http.StatusNotFound)
				errorCode := ErrorList{ErrorCode: 404, Message: param.ByName("idx") + "(Index) is not Found."}
				json.NewEncoder(writer).Encode(errorCode)
			}
		}
	})

	// 각 경로 별 GET, POST, PUT, DELETE, PATCH 메소드에 따른 처리 지정
	router.GET("/users", usersGetHandle)
	router.POST("/users", userPostHandle)

	router.GET("/users/:idx", userGetHandle)
	router.PUT("/users/:idx", userPutHandle)
	router.DELETE("/users/:idx", userDeleteHandle)
	router.PATCH("/users/:idx", userPatchHandle)

	// HTTP 포트 Listen
	http.ListenAndServe(":8080", router)
}
