import streamlit as st
from requests import request, Response, RequestException, put
from data import User
from chatbot import GET_DEFAULT_PROMPT, get_prompt

GET_SIGNUP_REQUESTS: str = "http://backend:8080/api/get/signup_request"
ACCEPT_SIGNUP_REQUEST: str = "http://backend:8080/api/update/signup_request"
REJECT_SIGNUP_REQUEST: str = "http://backend:8080/api/delete/signup_request"


GET_USER: str = "http://backend:8080/api/get/users"
PROMOTE_USER: str = "http://backend:8080/api/update/promote_user"
DELETE_USER: str = "http://backend:8080/api/delete/user"

UPDATE_DEFAULT_PROMPT: str = "http://backend:8080/api/update/default_prompt"

GET_MODELS: str = "http://backend:8080/api/get/models"
GET_SELECTED_MODEL: str = "http://backend:8080/api/get/current_model"
UPDATE_MODELS: str = "http://backend:8080/api/update/model_selection"

@st.fragment
def admin_dashboard(user: User):
    st.set_page_config(
        page_title="Admin Dashboard", 
        page_icon="./assets/Logo.png", 
        layout=None, 
        initial_sidebar_state=None, 
        menu_items=None
    )
    back, reload = st.columns(2)

    with back:
        if st.button(label="Back"):
            st.session_state["Admin Dashboard"] = False
            st.rerun()
    
    with reload:
        if st.button(label="Reload"):
            st.session_state["Admin Dashboard"] = True
            st.rerun()

    manage_signups(jwt=user.get_jwt())
    manage_users(jwt=user.get_jwt())

    llm_selection(jwt=user.get_jwt())
    update_default_prompt(jwt=user.get_jwt())

@st.fragment
def manage_signups(jwt: str):
    """
    Manages user signup requests through an admin interface.

    Requirements:
        - Admin privileges
        - Valid JWT authentication

    Behavior:d=None,
        - Displays pending signup requests in an expandable UI section
        - For each request, provides options to approve or reject
        - Approved users are added with regular user privileges
        - Shows toast notifications for approval/rejection actions

    UI Components:
        - Expandable container for signup requests
        - Column layout for each request (email display + action buttons)
        - Conditional display when no requests exist

    Side Effects:
        - Modifies user database on approval
        - Generates toast notifications
        - May make API calls to backend services

    Notes:
        - Currently uses mock data (signups list)
        - Actual backend integration not implemented
        - Buttons are uniquely keyed to prevent collisions
    """
    response: Response | None = execute_backend_operation(
        url=GET_SIGNUP_REQUESTS, 
        method="GET", 
        headers={"Authorization": jwt}, 
        json_payload=None,
        data=None
    )

    if response == None:
        st.stop()
    

    """
    Returns a List of Json-Objects

    {
        "Name": str,
        "Email": str,
    }
    """
    signups: list[dict[str, str]] | None = response.json()

    with st.expander("Sign Up requests"):
        if signups == None:
            st.write("No pending sign ups")
            return
        
        for index, request in enumerate(signups):
            text, approve, reject = st.columns([3, 1, 1])

            email: str = request["Email"]

            with text:
                st.write(email)
            
            with approve:
                if st.button(label="Accept", key=f"{index}_approve"):
                    response: Response | None = execute_backend_operation(
                        url=ACCEPT_SIGNUP_REQUEST,
                        method="PUT",
                        headers={"Authorization": jwt},
                        json_payload=None,
                        data=email
                    )

                    if response == None:
                        st.stop()
                    
                    if response.status_code == 200:
                        st.toast(f"Accepting {email}")
                    else:
                        st.error(response.content.decode("utf-8"))

                    st.rerun(scope="fragment")

            with reject:
                if st.button(label="Reject", key=f"{index}_reject"):
                    response: Response | None = execute_backend_operation(
                        url=REJECT_SIGNUP_REQUEST,
                        method="DELETE",
                        headers={"Authorization": jwt},
                        json_payload=None,
                        data=email
                    )

                    if response == None:
                        st.stop()

                    if response.status_code == 200:
                        st.toast(f"Rejecting {email}")
                    else:
                        st.error(response.content.decode("utf-8"))
                    
                    st.rerun(scope="fragment")

@st.fragment
def manage_users(jwt: str):
    """
    Manages existing users through an admin interface with promotion and deletion capabilities.

    Requirements:
        - Admin privileges
        - Valid JWT authentication

    Behavior:
        - Displays current users in an expandable UI section
        - Shows warning about irreversible actions
        - Provides options to promote users to admin or delete users
        - Shows toast notifications for completed actions

    UI Components:
        - Expandable container labeled "Manage Users"
        - Informational warning about action consequences
        - Column layout for each user (email display + action buttons)
        - Conditional display when no users exist

    Side Effects:
        - Modifies user roles on promotion
        - Removes users permanently on deletion
        - Generates toast notifications for actions
        - May make API calls to backend services

    Notes:
        - Uses mock data (users list) in current implementation
        - Backend integration not yet implemented
        - Promotion grants full admin privileges with no restrictions
        - Buttons are uniquely keyed to prevent UI collisions
        - All actions are irreversible as noted in UI warning
    """
    response: Response | None = execute_backend_operation(
        url=GET_USER,
        method="GET",
        headers={"Authorization": jwt},
        json_payload=None,
        data=None
    )


    if response == None:
        st.stop()
    

    """
    Returns a List of Json-Objects

    {
        "Email": str,
        "IsAdmin": bool,
    }
    """
    users: list[dict[str, str | bool]] | None = response.json()
    
    with st.expander(label="Manage Users"):
        if users == None:
            st.error("No  users")
            return

        st.info(
            "Promoting or deleting users is irreversible, proceed with caution. Promoted users (Admin) will no longer fall under any management jurristriction."
        )

        for index, user in enumerate(users):
            email, promote, delete = st.columns([3, 1, 1])

            with email:
                if user["IsAdmin"]:
                    st.write(f"⭐ {user["Email"]}")
                else:
                    st.write(user["Email"])
            
            with promote:
                if st.button(label="Promote", key=f"{index}_promote_user", disabled=user["IsAdmin"] == True):
                    user_email: str | bool | None = user.get("Email")
                    assert type(user_email) == str, "User email is not of type string"

                    response: Response | None = execute_backend_operation(
                        url=PROMOTE_USER,
                        method="PUT",
                        headers={"Authorization": jwt},
                        json_payload=None,
                        data=user_email
                    )

                    if response == None:
                        st.stop()
                        return

                    if response.status_code == 200:
                        st.toast(f"Promoted {user["Email"]}")
                    else:
                        st.error(response.content.decode("utf-8"))
                    st.rerun(scope="fragment")

            with delete: 
                if st.button(label="Delete", key=f"{index}_delete_user", disabled=len(users) == 1):
                    user_email: str | bool | None = user.get("Email")
                    assert type(user_email) == str, "User email is not of type string"

                    response: Response | None = execute_backend_operation(
                        url=DELETE_USER,
                        method="DELETE",
                        headers={"Authorization": jwt},
                        json_payload=None,
                        data=user_email
                    )


                    if response == None:
                        st.stop()
                        return
                    
                    if response.status_code == 200:
                        st.toast(f"Deleted {user["Email"]}")
                    else:
                        st.error(response.content.decode("utf-8"))
                    st.rerun(scope="fragment")

def llm_selection(jwt: str):
    """
    Fetches available LLM models from backend and allows user selection.
    
    Args:
        jwt: Authorization token for backend API
        
    Returns:
        str: The selected model name
        None: If no selection was made or if an error occurred
    """
    def update_model(model: str | None):
        """
        Helper function to update the selected LLM. This function is generally not expected to fail
        """
        if model == None:
            return
        
        try:
            response: Response = put(
                url=UPDATE_MODELS,
                headers={"Authorization": jwt}, 
                data=model
            )

            if response.status_code == 200:
                st.toast(f"Successfully update Model to {model}")
            else:
                st.error(f"Updating failed with: {response.content.decode("utf-8")}")
        except RequestException as e:
            st.error(f"Updating failed with: {e}")


    # Query available models
    response: Response | None = execute_backend_operation(
        url=GET_MODELS, 
        method="GET",
        headers={"Authorization": jwt},
        json_payload=None,
        data=None
    )

    # Query currently selected model
    response_current_model: Response | None = execute_backend_operation(
        url=GET_SELECTED_MODEL,
        method="GET",
        headers={"Authorization": jwt},
        json_payload=None,
        data=None
    )


    # Sanity check
    if response is None or response_current_model is None:
        st.error("Failed to fetch models from backend")
        return None

    try:
        models_data: dict[str, list[dict[str, str]]] = response.json()
        current_model: str = response_current_model.content.decode("utf-8")
        
        # Extract model names from the response
        if not isinstance(models_data, dict) or "models" not in models_data:
            st.error("Unexpected response format from backend")
            return None
            
        model_names: list[str] = [model["name"] for model in models_data["models"]]
        index: int = model_names.index(current_model)
        
        # Create the selectbox with actual models
        
        selected_model: str | None = st.selectbox(
            label="Select Backend AI Model",
            options=model_names,
            index=index,
            on_change=lambda: st.session_state.update({"Update_Model": True})
        )


        if st.session_state.get("Update_Model"):
            update_model(selected_model)
            st.session_state["Update_Model"] = None
        
    except (ValueError, KeyError) as e:
        st.error(f"Error processing models: {e}")
        return None

def update_default_prompt(jwt: str):
    """
    Ermöglicht die Manipulation des Standard-Prompts über eine Streamlit-Benutzeroberfläche.

    Diese Funktion ruft den aktuellen Standard-Prompt ab, ermöglicht dessen Bearbeitung 
    in einem Textbereich und sendet die Änderungen an ein Backend zur Aktualisierung. 
    Die Authentifizierung erfolgt über ein JSON Web Token (JWT).
    """
    import time


    with st.expander(label="Manage Default Prompt"):
        default_prompt: str | None = get_prompt(url=GET_DEFAULT_PROMPT, jwt=jwt)

        if default_prompt is None:
            st.error("Failed to load default prompt")
            return
        
        with st.container(border=True):
            st.write(default_prompt)
        

        new_prompt: str = st.text_area(
            label="Editor", 
            value=default_prompt, 
            on_change=lambda: st.session_state.update({"update_default_prompt": True})
        )


        if st.session_state.get("update_default_prompt"):
            st.session_state.update({"update_default_prompt": None})


            if response := execute_backend_operation(
                url=UPDATE_DEFAULT_PROMPT, 
                method="PUT",
                headers={"Authorization": jwt},
                json_payload=None,
                data=new_prompt
            ):
                if response.status_code != 200:
                    st.error(f"Failed to update default prompt with: {response.content.decode("utf-8")}")
                    time.sleep(0.8)
                else:
                    st.toast("Successfully update default prompt!")
            else:
                time.sleep(0.8)
                return


def execute_backend_operation(
        url: str, 
        method: str, 
        headers: dict[str, str], 
        json_payload: dict | None,
        data: bytes | str | None
    ) -> Response | None:
    """
    Executes a standardized HTTP request to the backend API.

    This function serves as the primary interface between the UI and backend services,
    handling all HTTP communication with error handling and logging.

    Args:
        url: The complete API endpoint URL to call
        method: HTTP method (GET, POST, PUT, DELETE, etc.)
        headers: Dictionary of HTTP headers to include (must include Authorization)
        json_payload: Optional JSON-serializable payload for the request

    Returns:
        Response: The successful HTTP response object if the request succeeds
        None: If the request fails (errors are displayed via st.error)

    Raises:
        RequestException: Propagates any underlying requests library exceptions
        ValueError: If required parameters are missing or invalid

    Side Effects:
        - Displays error messages via Streamlit's st.error on failure
        - May log errors to application monitoring systems

    Notes:
        - All backend calls should route through this centralized function
        - Callers must handle the None return case appropriately
        - Timeout configuration should be added in production
    """
    
    try:
        response: Response = request(url=url, method=method, headers=headers, json=json_payload, data=data)
        return response
    except RequestException as e:
        st.error(f"Failed to communicate with backend: {e}")
    return None