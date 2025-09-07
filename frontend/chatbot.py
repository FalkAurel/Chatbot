import streamlit as st
from data import Message, User, Prompt, Kind
from requests import put, post, get, request, RequestException, Response
from typing import Callable
from os import mkdir, remove
from pathlib import Path
from time import sleep

FILE_UPLOAD: str = "http://backend:8080/api/upload/file"
MESSAGE_UPLOAD: str = "http://backend:8080/api/upload/message"
MESSAGE_INFERENCE: str = "http://backend:8080/api/message/inference"

LOCAL_ONLY_UPDATE: str = "http://backend:8080/api/update/local_only"
ENABLE_LEGAL_LIBARY_UPDATE: str = "http://backend:8080/api/update/legal_libary"
GET_PROMPT: str = "http://backend:8080/api/get/prompt"
GET_DEFAULT_PROMPT: str = "http://backend:8080/api/get/default_prompt"
PROMPT_UPDATE: str = "http://backend:8080/api/update/prompt"

GET_CHAT_HISTORY: str = "http://backend:8080/api/get/history"
DELETE_CHAT_HISTORY: str = "http://backend:8080/api/delete/chat"

GET_DOCUMENTS: str = "http://backend:8080/api/get/documents"
DELETE_DOCUMENT: str = "http://backend:8080/api/delete/document"

TEMPORARY_FILE_PATH: Path = Path("./temporary_files")

ONLINE_SERVICE_PROVIDERS: list[tuple[str, str]] = [
    ("DeepSeek", "https://chat.deepseek.com/")
] 


def chatbot():
    st.set_page_config(
        page_title="MingBai", 
        page_icon="./assets/Logo.png", 
        layout=None, 
        initial_sidebar_state=None, 
        menu_items=None
    )
    user: object | None = st.session_state.get("user")

    if not isinstance(user, User):
        st.error("User is not set correctly")
        st.stop()

    with st.sidebar:
        st.title(f"Welcome {user.get_username()}")

        file = st.file_uploader(label="Provide context for your AI-Agent", type="pdf")
        if file:
            if not TEMPORARY_FILE_PATH.exists():
                mkdir(TEMPORARY_FILE_PATH)

            with open(file=TEMPORARY_FILE_PATH.joinpath(file.name), mode="wb") as f:
                f.write(file.getvalue())

            with open(file=TEMPORARY_FILE_PATH.joinpath(file.name), mode="rb") as f:
                try:
                    headers: dict[str, str] = {
                        "Title": TEMPORARY_FILE_PATH.joinpath(file.name).name,
                        "X-Filename": file.name,
                        "Authorization": user.get_jwt()
                    }
                    response: Response = post(url=FILE_UPLOAD, data=f, headers=headers)

                    if response.status_code != 200:
                        st.warning(f"File upload failed: {response.content.decode("utf-8")}")
                    else:
                        st.toast(f"Succesfully uploaded {file.name}")
                except RequestException as e:
                    st.error(f"Upload failed with: {e}")
                finally:
                    remove(TEMPORARY_FILE_PATH.joinpath(file.name))
                    file = None

        st.divider()

        if documents := get_documents(url=GET_DOCUMENTS, jwt=user.get_jwt()):
            with st.expander(label="Documents"):
                    for index, document in enumerate(documents):
                        name, delete = st.columns([3, 1])

                        with name:
                            st.write(document["OriginalName"])

                        with delete:
                            if st.button(label="Delete", key=f"{index}_delete_documents"):
                                try:
                                    response: Response = request(
                                        url=DELETE_DOCUMENT,
                                        method="DELETE",
                                        headers={"Authorization": user.get_jwt()},
                                        data=document["StorageName"]
                                    )

                                    if response.status_code != 200:
                                        st.error(f"Deletion failed: {response.content.decode("utf-8")}")
                                    else:
                                        st.toast(f"Deleted {document['OriginalName']}: {response.content.decode("utf-8")}")


                                    sleep(0.5)
                                    st.rerun()
                                except RequestException as e:
                                    st.error(f"Deletion failed: {e}")

            st.divider()
        st.write("AI-Chatbot Einstellungen")

        if st.button(label="Clear Chat"):
            try:
                response: Response = request(
                    url=DELETE_CHAT_HISTORY, 
                    method="DELETE",
                    headers={"Authorization": user.get_jwt()}
                )

                if response.status_code == 200:
                    st.rerun()
                else:
                    st.error(f"Clearing chat history failed: {response.content.decode()}")
            except RequestException as e:
                st.error(f"Clearing chat history failed: {e}")

        st.checkbox(
            label="Local only", 
            value=user.get_local(),
            key="local_only_checkbox",
            on_change=lambda: update(
                url=LOCAL_ONLY_UPDATE,
                resource_name="local_only",
                resource=not user.get_local(),
                jwt=user.get_jwt(),
                callback=user.set_local
            )
        )

        st.checkbox(
            label="Legal Library", 
            value=user.get_legal_libary(),
            key="legal_library_checkbox",
            on_change=lambda: update(
                url=ENABLE_LEGAL_LIBARY_UPDATE, 
                resource_name="legal_library", 
                resource=not user.get_legal_libary(),
                jwt=user.get_jwt(),
                callback=user.set_legal_libary
            ) 
        )

        # Preprompt Field
        st.checkbox(
            label="Deep Think",
            value=user.is_deep_think(),
            on_change= lambda: user.set_deep_think(not user.is_deep_think())
        )

        with st.expander(label="Go Online"):
            for label, link in ONLINE_SERVICE_PROVIDERS:
                st.link_button(label=label, url=link)


        current_prompt: str | None = get_prompt(url=GET_PROMPT, jwt=user.get_jwt())
        default_prompt: str | None = get_prompt(url=GET_DEFAULT_PROMPT, jwt=user.get_jwt())


        if current_prompt is None or default_prompt is None:
            st.error("Failed to load prompt")
        else:
            selected: str = st.selectbox(
                label="Preprompt",
                options=("Legal", "Other"),
                index=0 if current_prompt == default_prompt else 1,
                on_change= lambda: st.session_state.update({"Update": True})
            )

            if st.session_state.get("Update"):
                if selected == "Legal":
                    update_prompt(jwt=user.get_jwt(), prompt=None)
                    st.session_state["Update"] = None
                    st.rerun(scope="app")

                else:
                    input: str = st.text_area(label="Update Prompt", value="", on_change=lambda: st.session_state.update(
                            {"prompt_update": True}
                        )
                    )

                    if st.session_state.get("prompt_update"):
                        update_prompt(jwt=user.get_jwt(), prompt=input)
                        st.session_state.update({"prompt_update": None})
                        st.session_state["Update"] = None
                        st.rerun(scope="app")

            else:
                with st.expander(label="View Preprompt"):
                    st.write(current_prompt)


        if user.is_admin():
            st.divider()
            if st.button("Admin Dashboard"):
                st.session_state["Admin Dashboard"] = True
                st.session_state["user"] = user
                st.rerun()


    if isinstance(user, User):
        get_chat_history(user=user)

        st.session_state.update({"Initial_Load": 0})
        chat_interface(user=user)



def update(url: str, resource_name: str, resource: str | bool, jwt: str, callback: Callable):
    """
    Updates a given resource on the server.
    On success it will apply a callback to apply the state change to the `User`-Object
    """
    def update_server(url: str, resource_name: str, resource: str | bool) -> bool:
        """
        Tries to update the server with the new resource. On success it'll return True.
        It also displays status messages for errors and on success.
        """
        try:
            payload: dict[str, str | bool] = {resource_name: resource}
            response: Response = put(url=url, json=payload, headers={"Authorization": jwt})
            
            if response.status_code != 200:
                st.warning(f"Updating {resource_name} failed: {response.content.decode("utf-8")}")
                return False
            st.toast(f"Updating {resource_name} successful")
            return True
        except RequestException as e:
            st.error(f"Could not connect to the backend{e}")
            return False
        
    if st.rerun:
        if update_server(url=url, resource_name=resource_name, resource=resource):
            callback(resource)
    
    st.write()


# Sehr ineffizient, da wir die ganze UI rerendern
#@st.fragment
def chat_interface(user):
    # Display existing messages
    for message in user.get_messages():
        msg_writer = st.chat_message(str(message.get_kind()))
        msg_writer.write(message.get_message())
    
    # Verbesserte JavaScript-Lösung für Auto-Scroll
    scroll_js = """
    <script>
    function scrollToBottom() {
        setTimeout(function() {
            // Methode 1: Scroll zum Chat Input
            const textarea = document.querySelector('[data-testid="stChatInputTextArea"]');
            if (textarea) {
                textarea.scrollIntoView({ 
                    behavior: 'smooth', 
                    block: 'end',
                    inline: 'nearest'
                });
            }
            
            // Methode 2: Scroll zum letzten Chat-Element
            const chatMessages = document.querySelectorAll('[data-testid="chatAvatarIcon-user"], [data-testid="chatAvatarIcon-assistant"]');
            if (chatMessages.length > 0) {
                const lastMessage = chatMessages[chatMessages.length - 1];
                lastMessage.scrollIntoView({ 
                    behavior: 'smooth', 
                    block: 'end'
                });
            }
            
            // Methode 3: Scroll zum Ende der Seite
            window.scrollTo({
                top: document.body.scrollHeight,
                behavior: 'smooth'
            });
        }, 100);
        
        // Zusätzlicher Scroll nach längerer Wartezeit
        setTimeout(function() {
            window.scrollTo({
                top: document.body.scrollHeight,
                behavior: 'smooth'
            });
        }, 500);
    }
    
    // Sofortiger Scroll
    scrollToBottom();
    
    // Überwache DOM-Änderungen für neue Nachrichten
    const observer = new MutationObserver(function(mutations) {
        let shouldScroll = false;
        mutations.forEach(function(mutation) {
            if (mutation.type === 'childList' && mutation.addedNodes.length > 0) {
                for (let node of mutation.addedNodes) {
                    if (node.nodeType === 1 && node.querySelector) {
                        if (node.querySelector('[data-testid="chatAvatarIcon-user"]') || 
                            node.querySelector('[data-testid="chatAvatarIcon-assistant"]')) {
                            shouldScroll = true;
                            break;
                        }
                    }
                }
            }
        });
        
        if (shouldScroll) {
            scrollToBottom();
        }
    });
    
    observer.observe(document.body, {
        childList: true,
        subtree: true
    });
    
    // Cleanup nach 30 Sekunden
    setTimeout(function() {
        observer.disconnect();
    }, 30000);
    </script>
    """
    
    # JavaScript ausführen
    from streamlit.components.v1 import html
    html(scroll_js, height=0)
    
    # Chat Input
    if chat_input := st.chat_input("Your message"):
        message = Message(kind=Kind.User, message=chat_input)
        user.add_msg(message)
        
        st.session_state.update({"RefreshMessages": True})
        
        with st.spinner(text="Generating response...", show_time=True):
            if ai_response := prompt_ai(user=user, msg=message):
                user.add_msg(Message(kind=Kind.AI, message=ai_response.json()["response"]))
        
        if st.session_state.get("Initial_Load") == 0 or st.session_state.get("RefreshMessages"):
            st.session_state.update({"RefreshMessages": None})
            st.session_state.update({"Initial_Load": None})
            #st.rerun(scope="fragment")
            st.rerun()

def prompt_ai(user: User, msg: Message) -> Response | None:
    """
    Sends a prompt to the AI

    Very important: Simplify logic!
    """
    payload: dict[str, str | int] = {
        "kind": msg.get_kind().value,
        "message": msg.get_message() 
    }

    headers: dict[str, str] = {
        "Authorization": user.get_jwt(),
        "Deep_think": str(user.is_deep_think()) # "False" or "True" 
    }

    try:
        # Prompt AI
        ai_response: Response = get(url=MESSAGE_INFERENCE, json=payload, headers=headers)
        if ai_response.status_code != 200:
            st.warning(f"Prompting AI failed with: {ai_response.content.decode("utf-8")}", icon="⚠️")
            return
        # Update AI-Context and Database
        user_upload: Response = post(url=MESSAGE_UPLOAD, json=payload, headers=headers)
        if user_upload.status_code != 200:
            user.get_messages().pop()
            st.warning(f"Updating AI-Context failed with: {user_upload.content.decode("utf-8")}", icon="⚠️")
            return
        
        payload: dict[str, int | str] = {
            "kind": Kind.AI.value,
            "message": ai_response.json()["response"]
        }

        ai_upload: Response = post(url=MESSAGE_UPLOAD, json=payload, headers=headers)
        if ai_upload.status_code != 200:
            user.get_messages().pop()
            st.warning(f"Updating AI-Context failed with: {user_upload.content.decode("utf-8")}", icon="⚠️")
            return
        
        st.success("Update successful")
        return ai_response
    except RequestException as e:
        st.error(f"Sending request failed with: {e}")


def get_chat_history(user: User):
    """
    Updates the user-object with the most recent chat history
    Updating is not done incrementally, but rather in bulk

    This has following implications:
        - Built in synchronozation
        - More (unnecessary) data throughput
    """
    try:
        headers: dict[str, str] = {
            "Authorization": user.get_jwt()
        }
        response: Response = get(url=GET_CHAT_HISTORY, headers=headers)

        if response.status_code != 200:
            st.error(f"Failed to load previous chat history: {response.content.decode()}")
            st.stop()
        
        history: list[dict] = response.json()

        user.reset_history()
        for message in history:
            user.add_msg(Message(kind=Kind(message["kind"]), message=message["message"] ))
    except RequestException as e:
        st.error(f"Failed to connect to backend: {e}")
        st.stop()


def get_documents(url: str, jwt: str) -> list[dict[str, str]] | None:
    """
    Retrieves the documents for a given user.

    If the request does not fail, it will return a list of JSON-Objects:

    {
        "OriginalName": str
        "StorageName": str
    }

    If the request fails, it will return None
    """
    try:
        response: Response = get(url=url, headers={"Authorization": jwt})

        if response.status_code != 200:
            st.error(f"Request failed with: {response.content.decode("utf-8")}")
            return None
        
        documents: list[dict[str, str]] = response.json()
        return documents
    except RequestException as e:
        st.error(f"Failed to send with {e}")
        return None
    
def get_prompt(url: str, jwt: str) -> str | None:
    try:
        response: Response = get(url=url, headers={"Authorization": jwt})
        payload: str = response.content.decode("utf-8")

        if response.status_code != 200:
            st.error(f"Failed to load prompt: {payload}")
            return None
        return payload
    except RequestException as e:
        st.error(f"Failed to load prompt: {e}")
        return None
    
def update_prompt(jwt: str, prompt: str | None):
    try:
        response: Response = put(url=PROMPT_UPDATE, headers={"Authorization": jwt}, data=prompt)
        payload: str = response.content.decode("utf-8")

        if response.status_code != 200:
            st.error(f"Failed to load prompt: {payload}")
            return None
        st.toast("Successfully updated prompt")
    except RequestException as e:
        st.error(f"Failed to load prompt: {e}")
        return None