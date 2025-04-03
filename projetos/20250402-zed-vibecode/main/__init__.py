from fastapi import FastAPI, APIRouter, HTTPException, Depends, status
from fastapi.responses import JSONResponse
from typing import Dict, List, Union, Optional, Any, Tuple
from pydantic import BaseModel
import uvicorn
import os
import logging
from sqlalchemy import create_engine, Column, Integer, String, Boolean, inspect
from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy.orm import sessionmaker, Session
from sqlalchemy.exc import IntegrityError, SQLAlchemyError
from enum import Enum, auto

app = FastAPI(title="User API with SQLite")

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

# SQLAlchemy setup
SQLALCHEMY_DATABASE_URL = "sqlite:///./users.db"
engine = create_engine(SQLALCHEMY_DATABASE_URL, connect_args={"check_same_thread": False})
SessionLocal = sessionmaker(autocommit=False, autoflush=False, bind=engine)
Base = declarative_base()

# Exception type enum
class APIErrorType(Enum):
    # Common errors
    VALIDATION_ERROR = (status.HTTP_400_BAD_REQUEST, "Invalid data provided")
    NOT_FOUND = (status.HTTP_404_NOT_FOUND, "Resource not found")
    INTERNAL_ERROR = (status.HTTP_500_INTERNAL_SERVER_ERROR, "An unexpected error occurred")

    # User specific errors
    USER_NOT_FOUND = (status.HTTP_404_NOT_FOUND, "User not found")
    USER_ALREADY_EXISTS = (status.HTTP_400_BAD_REQUEST, "Username or email already exists")
    USER_UPDATE_FAILED = (status.HTTP_400_BAD_REQUEST, "Failed to update user")
    USER_CREATE_FAILED = (status.HTTP_400_BAD_REQUEST, "Failed to create user")
    USER_DELETE_FAILED = (status.HTTP_400_BAD_REQUEST, "Failed to delete user")
    EMAIL_EXISTS = (status.HTTP_400_BAD_REQUEST, "Email already exists")

    # Repository specific errors
    USERNAME_EMAIL_EXISTS = (status.HTTP_400_BAD_REQUEST, "Username or email already exists")
    FAILED_CREATE_USER = (status.HTTP_400_BAD_REQUEST, "Failed to create user")
    FAILED_FETCH_USERS = (status.HTTP_500_INTERNAL_SERVER_ERROR, "Failed to fetch users")
    FAILED_FETCH_USER_BY_ID = (status.HTTP_500_INTERNAL_SERVER_ERROR, "Failed to fetch user by ID")
    FAILED_FETCH_USER_BY_USERNAME = (status.HTTP_500_INTERNAL_SERVER_ERROR, "Failed to fetch user by username")
    FAILED_FETCH_USER_BY_IDENTIFIER = (status.HTTP_500_INTERNAL_SERVER_ERROR, "Failed to fetch user by identifier")
    EMAIL_ALREADY_EXISTS = (status.HTTP_400_BAD_REQUEST, "Email already exists")
    FAILED_UPDATE_USER = (status.HTTP_400_BAD_REQUEST, "Failed to update user")
    FAILED_DELETE_USER = (status.HTTP_400_BAD_REQUEST, "Failed to delete user")

    # Database errors
    DB_FETCH_ERROR = (status.HTTP_500_INTERNAL_SERVER_ERROR, "Failed to fetch data")
    DB_UPDATE_ERROR = (status.HTTP_500_INTERNAL_SERVER_ERROR, "Failed to update data")
    DB_INIT_ERROR = (status.HTTP_500_INTERNAL_SERVER_ERROR, "Failed to initialize database")

    def __init__(self, status_code: int, detail: str):
        self.status_code = status_code
        self.detail = detail

    def exception(self, custom_detail: str = None) -> HTTPException:
        """Return an HTTPException with the status code and detail message instead of raising it directly"""
        detail = custom_detail if custom_detail else self.detail
        return HTTPException(status_code=self.status_code, detail=detail)

# SQLAlchemy Model
class UserModel(Base):
    __tablename__ = "users"

    id = Column(Integer, primary_key=True, index=True)
    username = Column(String, unique=True, nullable=False)
    email = Column(String, unique=True, nullable=False)
    full_name = Column(String, nullable=True)
    active = Column(Boolean, default=True)

# Dependency to get the database session
def get_db():
    db = SessionLocal()
    try:
        yield db
    finally:
        db.close()

# Initialize database
def init_db():
    """Initialize database and create tables if they don't exist"""
    try:
        if not inspect(engine).has_table("users"):
            Base.metadata.create_all(bind=engine)
            logger.info("Database initialized successfully")
        else:
            logger.info("Database already exists")
    except SQLAlchemyError as e:
        logger.error(f"Error initializing database: {e}")
        raise APIErrorType.DB_INIT_ERROR.exception(f"Database initialization failed: {str(e)}")

# Pydantic models for API
class UserBase(BaseModel):
    username: str
    email: str
    full_name: Optional[str] = None
    active: bool = True

class UserCreate(UserBase):
    pass

class User(UserBase):
    id: int

    class Config:
        orm_mode = True

class UserPatch(BaseModel):
    email: Optional[str] = None
    full_name: Optional[str] = None
    active: Optional[bool] = None

# Repository pattern implementation
class UserRepository:
    def __init__(self, db: Session):
        self.db = db

    def create(self, user: UserCreate) -> User:
        """Create a new user"""
        db_user = UserModel(
            username=user.username,
            email=user.email,
            full_name=user.full_name,
            active=user.active
        )
        try:
            self.db.add(db_user)
            self.db.commit()
            self.db.refresh(db_user)
            return db_user
        except IntegrityError:
            self.db.rollback()
            raise ValueError(APIErrorType.USERNAME_EMAIL_EXISTS.detail)
        except SQLAlchemyError as e:
            self.db.rollback()
            logger.error(f"Error creating user: {e}")
            raise ValueError(APIErrorType.FAILED_CREATE_USER.detail)

    def get_all(self) -> List[UserModel]:
        """Get all users"""
        try:
            return self.db.query(UserModel).all()
        except SQLAlchemyError as e:
            logger.error(f"Error fetching all users: {e}")
            raise ValueError(APIErrorType.FAILED_FETCH_USERS.detail)

    def get_by_id(self, user_id: int) -> Optional[UserModel]:
        """Get a user by ID"""
        try:
            return self.db.query(UserModel).filter(UserModel.id == user_id).first()
        except SQLAlchemyError as e:
            logger.error(f"Error fetching user by ID: {e}")
            raise ValueError(APIErrorType.FAILED_FETCH_USER_BY_ID.detail)

    def get_by_username(self, username: str) -> Optional[UserModel]:
        """Get a user by username"""
        try:
            return self.db.query(UserModel).filter(UserModel.username == username).first()
        except SQLAlchemyError as e:
            logger.error(f"Error fetching user by username: {e}")
            raise ValueError(APIErrorType.FAILED_FETCH_USER_BY_USERNAME.detail)

    def get_by_identifier(self, identifier: Any) -> Optional[UserModel]:
        """Get a user by ID or username"""
        try:
            user_id = int(identifier)
            return self.get_by_id(user_id)
        except ValueError:
            return self.get_by_username(identifier)
        except SQLAlchemyError as e:
            logger.error(f"Error fetching user by identifier: {e}")
            raise ValueError(APIErrorType.FAILED_FETCH_USER_BY_IDENTIFIER.detail)

    def update(self, user_id: int, user_data: dict) -> Optional[UserModel]:
        """Update a user"""
        db_user = self.get_by_id(user_id)
        if not db_user:
            return None

        try:
            for key, value in user_data.items():
                if value is not None:
                    setattr(db_user, key, value)

            self.db.commit()
            self.db.refresh(db_user)
            return db_user
        except IntegrityError:
            self.db.rollback()
            raise ValueError(APIErrorType.EMAIL_ALREADY_EXISTS.detail)
        except SQLAlchemyError as e:
            self.db.rollback()
            logger.error(f"Error updating user: {e}")
            raise ValueError(APIErrorType.FAILED_UPDATE_USER.detail)

    def delete(self, user_id: int) -> bool:
        """Delete a user"""
        db_user = self.get_by_id(user_id)
        if not db_user:
            return False

        try:
            self.db.delete(db_user)
            self.db.commit()
            return True
        except SQLAlchemyError as e:
            self.db.rollback()
            logger.error(f"Error deleting user: {e}")
            raise ValueError(APIErrorType.FAILED_DELETE_USER.detail)

# Create user router
user_router = APIRouter(prefix="/users", tags=["users"])

# User CRUD operations
@user_router.post("/", response_model=User, status_code=status.HTTP_201_CREATED)
def create_user(user: UserCreate, db: Session = Depends(get_db)):
    """Create a new user"""
    try:
        repo = UserRepository(db)
        return repo.create(user)
    except ValueError as e:
        if str(e) == APIErrorType.USERNAME_EMAIL_EXISTS.detail:
            raise APIErrorType.USER_ALREADY_EXISTS.exception()
        else:
            raise APIErrorType.USER_CREATE_FAILED.exception(str(e))
    except SQLAlchemyError as e:
        logger.error(f"Database error creating user: {e}")
        raise APIErrorType.DB_UPDATE_ERROR.exception(str(e))
    except Exception as e:
        logger.error(f"Unexpected error creating user: {e}")
        raise APIErrorType.INTERNAL_ERROR.exception(str(e))

@user_router.get("/", response_model=List[User])
def get_users(db: Session = Depends(get_db)):
    """Get all users"""
    try:
        repo = UserRepository(db)
        return repo.get_all()
    except ValueError as e:
        logger.error(f"Value error fetching users: {e}")
        raise APIErrorType.DB_FETCH_ERROR.exception(str(e))
    except SQLAlchemyError as e:
        logger.error(f"Database error fetching users: {e}")
        raise APIErrorType.DB_FETCH_ERROR.exception(str(e))
    except Exception as e:
        logger.error(f"Unexpected error fetching users: {e}")
        raise APIErrorType.INTERNAL_ERROR.exception(str(e))

@user_router.get("/{user_identifier}", response_model=User)
def get_user(user_identifier: Any, db: Session = Depends(get_db)):
    """Get a user by ID or username"""
    try:
        repo = UserRepository(db)
        user = repo.get_by_identifier(user_identifier)
        if user is None:
            raise APIErrorType.USER_NOT_FOUND.exception()
        return user
    except ValueError as e:
        logger.error(f"Value error fetching user: {e}")
        raise APIErrorType.DB_FETCH_ERROR.exception(str(e))
    except SQLAlchemyError as e:
        logger.error(f"Database error fetching user: {e}")
        raise APIErrorType.DB_FETCH_ERROR.exception(str(e))
    except HTTPException:
        # Re-raise HTTP exceptions (like USER_NOT_FOUND)
        raise
    except Exception as e:
        logger.error(f"Unexpected error fetching user: {e}")
        raise APIErrorType.INTERNAL_ERROR.exception(str(e))

@user_router.patch("/{user_identifier}", response_model=User)
def patch_user(user_identifier: Any, user_update: UserPatch, db: Session = Depends(get_db)):
    """Partially update a user's email, full_name, or active status"""
    try:
        repo = UserRepository(db)

        # Find the user
        user = repo.get_by_identifier(user_identifier)
        if user is None:
            raise APIErrorType.USER_NOT_FOUND.exception()

        # Extract non-None values from the update
        update_data = user_update.dict(exclude_unset=True, exclude_none=True)

        # If nothing to update, return the current user
        if not update_data:
            return user

        # Update the user
        updated_user = repo.update(user.id, update_data)
        return updated_user
    except ValueError as e:
        if str(e) == APIErrorType.EMAIL_ALREADY_EXISTS.detail:
            raise APIErrorType.EMAIL_EXISTS.exception()
        else:
            raise APIErrorType.USER_UPDATE_FAILED.exception(str(e))
    except SQLAlchemyError as e:
        logger.error(f"Database error patching user: {e}")
        raise APIErrorType.DB_UPDATE_ERROR.exception(str(e))
    except HTTPException:
        # Re-raise HTTP exceptions (like USER_NOT_FOUND)
        raise
    except Exception as e:
        logger.error(f"Unexpected error patching user: {e}")
        raise APIErrorType.INTERNAL_ERROR.exception(str(e))

@user_router.delete("/{user_identifier}", status_code=status.HTTP_204_NO_CONTENT)
def delete_user(user_identifier: Any, db: Session = Depends(get_db)):
    """Delete a user by ID or username"""
    try:
        repo = UserRepository(db)

        # Find the user
        user = repo.get_by_identifier(user_identifier)
        if user is None:
            raise APIErrorType.USER_NOT_FOUND.exception()

        # Delete the user
        repo.delete(user.id)
        return None
    except ValueError as e:
        raise APIErrorType.USER_DELETE_FAILED.exception(str(e))
    except SQLAlchemyError as e:
        logger.error(f"Database error deleting user: {e}")
        raise APIErrorType.DB_UPDATE_ERROR.exception(str(e))
    except HTTPException:
        # Re-raise HTTP exceptions (like USER_NOT_FOUND)
        raise
    except Exception as e:
        logger.error(f"Unexpected error deleting user: {e}")
        raise APIErrorType.INTERNAL_ERROR.exception(str(e))

# Create versioned API for compatibility
v1_router = APIRouter(prefix="/v1")
v2_router = APIRouter(prefix="/v2")

# V1 routes
@v1_router.get("/hello")
def hello_v1() -> Dict[str, str]:
    """Hello world endpoint for API v1"""
    return {"message": "Hello World from API v1!"}

@v1_router.get("/hello/{name}")
def hello_name_v1(name: str) -> Dict[str, str]:
    """Hello with name parameter for API v1"""
    return {"message": f"Hello {name} from API v1!"}

# V2 routes (could have different implementation or additional features)
@v2_router.get("/hello")
def hello_v2() -> Dict[str, str]:
    """Hello world endpoint for API v2"""
    return {"message": "Hello World from API v2!", "version": "2.0"}

@v2_router.get("/hello/{name}")
def hello_name_v2(name: str) -> Dict[str, str]:
    """Hello with name parameter for API v2"""
    return {"message": f"Hello {name} from API v2!", "version": "2.0"}

# Root endpoint
@app.get("/")
def root() -> Dict[str, Union[str, List[str]]]:
    """Root endpoint providing API information"""
    return {
        "app_name": "User API with SQLite",
        "available_endpoints": ["/users", "/v1", "/v2"],
        "documentation": "/docs"
    }

# Include all routers in the main app
app.include_router(user_router)
app.include_router(v1_router)
app.include_router(v2_router)

# Run with: uvicorn main:app --reload
def main():
    """Run the FastAPI application with uvicorn server"""
    try:
        # Initialize the database
        init_db()

        # Get port from environment variable or use 5001 as default
        port = int(os.environ.get("PORT", 5001))
        logger.info(f"Starting server on port {port}...")
        uvicorn.run("main:app", host="0.0.0.0", port=port, reload=True)
    except ValueError as e:
        logger.error(f"Invalid port configuration: {e}")
        sys.exit(1)
    except Exception as e:
        logger.error(f"Failed to start application: {e}")
        sys.exit(1)

if __name__ == "__main__":
    import sys
    main()
